// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"crypto/tls"
	"crypto/x509"
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
	SONIC_RESTCONF_MACTABLE = "/restconf/data/openconfig-network-instance:network-instances/network-instance=default/fdb/mac-table/entries"
	SONIC_RESTCONF_LLDP     = "/restconf/data/openconfig-lldp:lldp/interfaces"
)

type Sonic struct {
	username string
	password string
	baseUrl  string
	client   *http.Client
}

type sonicMacTable struct {
	Root struct {
		Entry []sonicMacTableEntry `json:"entry"`
	} `json:"openconfig-network-instance:entries"`
}

type sonicMacTableEntry struct {
	MacAddress string `json:"mac-address"`
	Vlan       int    `json:"vlan"`
	State      struct {
		EntryType  string `json:"entry-type"`
		MacAddress string `json:"mac-address"`
		Vlan       int    `json:"vlan"`
	} `json:"state"`
	Interface struct {
		InterfaceRef struct {
			State struct {
				Interface string `json:"interface"`
			} `json:"state"`
		} `json:"interface-ref"`
	} `json:"interface"`
}

func NewSonic(baseUrl, user, password, cacert string, insecure bool) (*Sonic, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}}

	pem, err := os.ReadFile(cacert)
	if err == nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("Failed to read cacert: %s", cacert)
		}

		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: certPool, InsecureSkipVerify: false}}
	}

	d := &Sonic{
		username: user,
		password: password,
		baseUrl:  "https://" + baseUrl,
		client:   &http.Client{Timeout: time.Second * 20, Transport: tr},
	}

	return d, nil
}

func (d *Sonic) URL(resource string) string {
	return fmt.Sprintf("%s%s", d.baseUrl, resource)
}

func (d *Sonic) getRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if d.username != "" && d.password != "" {
		req.SetBasicAuth(d.username, d.password)
	}

	return req, nil
}

func (d *Sonic) GetMACTable() (model.MACTable, error) {
	url := d.URL(SONIC_RESTCONF_MACTABLE)
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
		return nil, fmt.Errorf("failed to fetch mac table with HTTP status code: %d", res.StatusCode)
	}

	rawJson, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	log.Debugf("Sonic json response: %s", rawJson)

	var sMACTable *sonicMacTable
	err = json.Unmarshal(rawJson, &sMACTable)
	if err != nil {
		return nil, err
	}

	macTable := make(model.MACTable, 0)

	for _, entry := range sMACTable.Root.Entry {
		// Parse port number from interface.
		// Format is: Eth1/16 (in standard naming mode)
		iface := entry.Interface.InterfaceRef.State.Interface
		if !strings.HasPrefix(iface, "Eth1/") {
			continue
		}
		portStr := strings.Replace(iface, "Eth1/", "", 1)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Debugf("failed to parse mac address table port on interface: %s", iface)
			continue
		}

		mac, err := net.ParseMAC(entry.MacAddress)
		if err != nil {
			log.Errorf("Invalid mac address entry %s: %v", entry.MacAddress, err)
			continue
		}

		macTable[entry.MacAddress] = &model.MACTableEntry{
			Ifname: iface,
			Port:   port,
			VLAN:   strconv.Itoa(entry.Vlan),
			Type:   entry.State.EntryType,
			MAC:    mac,
		}
	}

	log.Infof("Received %d entries", len(macTable))
	return macTable, nil

}

type sonicLLDP struct {
	Root struct {
		Interface []struct {
			Name      string `json:"name"`
			Neighbors struct {
				Neighbor []struct {
					Id    string `json:"id"`
					State struct {
						ChassisId         string `json:"chassis-id"`
						ChassisidType     string `json:"chassis-id-type"`
						Id                string
						ManagementAddress string `json:"management-address"`
						PortDescription   string `json:"port-description"`
						PortId            string `json:"port-id"`
						PortIdType        string `json:"port-id-type"`
						SystemDescription string `json:"system-description"`
						SystemName        string `json:"system-name"`
						Ttl               int
					} `json:"state"`
				} `json:"neighbor"`
			} `json:"neighbors"`
		} `json:"interface"`
	} `json:"openconfig-lldp:interfaces"`
}

func (d *Sonic) GetInterfaceStatus() (model.InterfaceTable, error) {
	return nil, errors.New("Interface Status not supported on SONiC")
}

// TODO:
func (d *Sonic) GetLLDPNeighbors() (model.LLDPNeighbors, error) {
	url := d.URL(SONIC_RESTCONF_LLDP)
	log.Infof("Requesting LLDP info: %s", url)

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
		return nil, fmt.Errorf("failed to fetch mac table with HTTP status code: %d", res.StatusCode)
	}

	rawJson, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var lldpRaw *sonicLLDP
	err = json.Unmarshal(rawJson, &lldpRaw)
	if err != nil {
		return nil, err
	}

	o := make(model.LLDPNeighbors, 0)

	for _, iface := range lldpRaw.Root.Interface {
		for _, n := range iface.Neighbors.Neighbor {
			o[iface.Name] = &model.LLDP{
				PortName:          iface.Name,
				ChassisId:         n.State.ChassisId,
				ChassisIdType:     n.State.ChassisidType,
				ManagementAddress: n.State.ManagementAddress,
				PortDescription:   n.State.PortDescription,
				PortId:            n.State.PortId,
				PortIdType:        n.State.PortIdType,
				SystemDescription: n.State.SystemDescription,
				SystemName:        n.State.SystemName,
			}
		}
	}

	return o, nil
	// return nil, errors.New("LLDPNeighbors not supported on SONiC")
}
