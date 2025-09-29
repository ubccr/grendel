// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ubccr/grendel/pkg/model"
)

const (
	DELLOS10_RESTCONF_MACTABLE   = "/restconf/data/dell-l2-mac:oper-params"
	DELLOS10_RESTCONF_Interfaces = "/restconf/data/interfaces-state/interface"
)

type DellOS10 struct {
	endpoint string
	user     string
	password string
	client   *http.Client
}

type dellMacTable struct {
	DynamicCount int                  `json:"dynamic-mac-count`
	StaticCount  int                  `json:"static-mac-count`
	Entries      []*dellMacTableEntry `json:"fwd-table"`
}

type dellMacTableEntry struct {
	PortIndex int    `json:"dot1d-port-index"`
	Type      string `json:"entry-type"`
	Ifname    string `json:"if-name"`
	MAC       string `json:"mac-addr"`
	Status    string `json:"status"`
	VLAN      string `json:"vlan"`
}

type dellIetfInterfaces struct {
	Interfaces []dellIetfInterface `json:"ietf-interfaces:interface"`
}

type dellIetfInterface struct {
	Name                    string                  `json:"name"`
	DellLldpRemNeighborInfo dellLldpRemNeighborInfo `json:"dell-lldp:lldp-rem-neighbor-info"`
	DellLldpRemMgmtAddr     dellLldpRemMgmtAddr     `json:"dell-lldp:rem-mgmt-addr"`
}
type dellLldpRemNeighborInfo struct {
	Info []struct {
		RemLldpChassisId        string `json:"rem-lldp-chassis-id"`
		RemLldpChassisIdSubtype string `json:"rem-lldp-chassis-id-subtype"`
		RemLldpPortId           string `json:"rem-lldp-port-id"`
		RemLldpPortSubtype      string `json:"rem-lldp-port-subtype"`
		RemSystemName           string `json:"rem-system-name"`
		RemSystemDesc           string `json:"rem-system-desc"`
		RemPortDesc             string `json:"rem-port-desc"`
	} `json:"info"`
}
type dellLldpRemMgmtAddr struct {
	Info []struct {
		RemMgmtAddress     string `json:"rem-mgmt-address"`
		RemMgmtAddrSubType string `json:"rem-mgmt-addr-sub-type"`
	}
}

type dellRestconfError struct {
	AppTag  string `json:"error-app-tag"`
	Message string `json:"error-message"`
	Tag     string `json:"error-tag"`
	Type    string `json:"type"`
}

func NewDellOS10(endpoint, user, password, cacert string, insecure bool) (*DellOS10, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}}

	pem, err := os.ReadFile(cacert)
	if err == nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("Failed to read cacert: %s", cacert)
		}

		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: certPool, InsecureSkipVerify: false}}
	}

	d := &DellOS10{
		user:     user,
		password: password,
		endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   &http.Client{Timeout: time.Second * 20, Transport: tr},
	}

	return d, nil
}

func (d *DellOS10) URL(resource string) string {
	return fmt.Sprintf("%s%s", d.endpoint, resource)
}

func (d *DellOS10) getRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if d.user != "" && d.password != "" {
		req.SetBasicAuth(d.user, d.password)
	}

	return req, nil
}

func (d *DellOS10) GetMACTable() (model.MACTable, error) {
	url := d.URL(DELLOS10_RESTCONF_MACTABLE)
	log.Infof("Requesting MAC table: %s", url)

	req, err := d.getRequest(url)
	if err != nil {
		return nil, err
	}
	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 500 {
		return nil, fmt.Errorf("Failed to fetch mac table with HTTP status code: %d", res.StatusCode)
	}

	rawJson, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	log.Debugf("DELLOS10 json response: %s", rawJson)

	var dmacTable map[string]*dellMacTable
	err = json.Unmarshal(rawJson, &dmacTable)
	if err != nil {
		return nil, err
	}

	if rec, ok := dmacTable["dell-l2-mac:oper-params"]; ok {
		macTable := make(model.MACTable, 0)

		for _, entry := range rec.Entries {
			// Parse port number from interface.
			// Format is: ethernet node/slot/port[:subport]
			parts := strings.Split(entry.Ifname, "/")
			if len(parts) != 3 || !strings.HasPrefix(parts[0], "ethernet") {
				log.Debugf("Invalid interface entry: %s", entry.Ifname)
				continue
			}

			port, err := strconv.Atoi(parts[2])
			if err != nil {
				log.Debugf("Invalid interface entry port number not a number: %s", entry.Ifname)
				continue
			}

			mac, err := net.ParseMAC(entry.MAC)
			if err != nil {
				log.Errorf("Invalid mac address entry %s: %v", entry.MAC, err)
				continue
			}

			macTable[entry.MAC] = &model.MACTableEntry{
				Ifname: entry.Ifname,
				Port:   port,
				VLAN:   entry.VLAN,
				Type:   entry.Type,
				MAC:    mac,
			}
		}

		log.Infof("Received %d entries", len(macTable))
		return macTable, nil
	}

	var derr map[string]map[string][]*dellRestconfError
	err = json.Unmarshal(rawJson, &derr)
	if err != nil {
		return nil, err
	}

	if erec, ok := derr["ietf-restconf:errors"]; ok {
		if rec, ok := erec["error"]; ok {
			if len(rec) > 0 {
				return nil, fmt.Errorf("Failed to fetch mac table: %s - %s", rec[0].Tag, rec[0].Message)
			}
		}
	}

	return nil, errors.New("Failed to fetch mac table, unknown error")
}

func (d *DellOS10) GetLLDPNeighbors() (model.LLDPNeighbors, error) {
	url := d.URL(DELLOS10_RESTCONF_Interfaces)
	log.Infof("Requesting interfaces for LLDP info: %s", url)

	req, err := d.getRequest(url)
	if err != nil {
		return nil, err
	}
	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 500 {
		return nil, fmt.Errorf("Failed to fetch interfaces with HTTP status code: %d", res.StatusCode)
	}

	rawJson, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// log.Debugf("DELLOS10 json response: %s", rawJson)

	var dIetfInterfaces dellIetfInterfaces
	err = json.Unmarshal(rawJson, &dIetfInterfaces)
	if err != nil {
		return nil, err
	}
	LldpTable := make(model.LLDPNeighbors, 0)

	for _, iface := range dIetfInterfaces.Interfaces {
		if len(iface.DellLldpRemNeighborInfo.Info) < 1 || len(iface.DellLldpRemMgmtAddr.Info) < 1 {
			continue
		}

		neighborInfo := iface.DellLldpRemNeighborInfo.Info[0]
		mgmtAddrInfo := iface.DellLldpRemMgmtAddr.Info[0]

		var chassisId string
		chassisIdType := neighborInfo.RemLldpChassisIdSubtype
		if chassisIdType == "mac-address" {
			// Dell encodes rem-lldp-chassis-id in base64 -> binary
			bChassisId, err := base64.StdEncoding.DecodeString(neighborInfo.RemLldpChassisId)
			if err != nil {
				log.Warnf("failed to decode dell-lldp-rem-neighbor-info for port: %s, %s", iface.Name, err)
			}

			sChassisId := fmt.Sprintf("%x", bChassisId)

			for idx, x := range sChassisId {
				chassisId += string(x)
				if idx%2 != 0 && idx < (len(sChassisId)-1) {
					chassisId += ":"
				}
			}
			chassisIdType = "MAC_ADDRESS"
		}
		mgmtAddr, err := base64.StdEncoding.DecodeString(mgmtAddrInfo.RemMgmtAddress)
		if err != nil {
			log.Warnf("failed to decode dell-lldp-rem-mgmt-address for port: %s", iface.Name)
			continue
		}
		portId, err := base64.StdEncoding.DecodeString(neighborInfo.RemLldpPortId)
		if err != nil {
			log.Warnf("failed to decode dell-lldp-rem-port-id for port: %s", iface.Name)
			continue
		}

		LldpTable[iface.Name] = &model.LLDP{
			PortName:          iface.Name,
			ChassisIdType:     chassisIdType,
			ChassisId:         chassisId,
			SystemName:        neighborInfo.RemSystemName,
			SystemDescription: neighborInfo.RemSystemDesc,
			ManagementAddress: string(mgmtAddr),
			PortDescription:   neighborInfo.RemPortDesc,
			PortId:            string(portId),
			PortIdType:        neighborInfo.RemLldpPortSubtype,
		}
	}

	log.Infof("Received %d entries", len(LldpTable))
	return LldpTable, nil

	// var derr map[string]map[string][]*dellRestconfError
	// err = json.Unmarshal(rawJson, &derr)
	// if err != nil {
	// 	return nil, err
	// }

	// if erec, ok := derr["ietf-restconf:errors"]; ok {
	// 	if rec, ok := erec["error"]; ok {
	// 		if len(rec) > 0 {
	// 			return nil, fmt.Errorf("Failed to fetch mac table: %s - %s", rec[0].Tag, rec[0].Message)
	// 		}
	// 	}
	// }

	// return nil, errors.New("Failed to fetch mac table, unknown error")
	// return nil, errors.New("LLDPNeighbors not supported on Dell OS10")
}

func (d *DellOS10) GetInterfaceStatus() (model.InterfaceTable, error) {
	return nil, errors.New("Interface Status not supported on Dell OS10")
}
