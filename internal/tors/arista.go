// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"net"
	"strconv"
	"strings"

	"github.com/aristanetworks/goeapi"
	"github.com/aristanetworks/goeapi/module"
)

const (
	ARISTA_MACTABLE = "show mac address-table"
)

type Arista struct {
	client *goeapi.Node
}

func NewArista(host string, username, password string) (*Arista, error) {
	node, err := goeapi.Connect("https", host, username, password, 443)
	if err != nil {
		return nil, err
	}

	return &Arista{client: node}, nil
}

func (a *Arista) GetInterfaceStatus() (InterfaceTable, error) {
	interfaceTable := make(InterfaceTable, 0)
	show := module.Show(a.client)
	statusRes := show.ShowInterfaces()

	for k, v := range statusRes.Interfaces {
		mac, err := net.ParseMAC(v.PhysicalAddress)
		if err != nil {
			log.Error(err)
		}
		portStr := strings.Replace(k, "Ethernet", "", 1)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		interfaceTable[port] = &InterfaceStatus{
			Name:                v.Name,
			LineProtocolStatus:  v.LineProtocolStatus,
			InterfaceStatus:     v.InterfaceStatus,
			PhysicalAddress:     mac,
			Description:         v.Description,
			Hardware:            v.Hardware,
			Bandwidth:           v.Bandwidth,
			MTU:                 v.Mtu,
			InterfaceMembership: v.InterfaceMembership,
		}
	}
	return interfaceTable, nil
}

func (a *Arista) GetMACTable() (MACTable, error) {
	macTable := make(MACTable, 0)
	show := module.Show(a.client)
	macRes, err := show.ShowMACAddressTable()
	if err != nil {
		return nil, err
	}

	// Should we also loop over the Multicast table??
	for _, m := range macRes.UnicastTable.TableEntries {
		if !strings.HasPrefix(m.Interface, "Ethernet") {
			continue
		}
		portStr := strings.Replace(m.Interface, "Ethernet", "", 1)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Debugf("failed to parse mac address table port on interface: %s", m.Interface)
			continue
		}
		mac, err := net.ParseMAC(m.MACAddress)
		if err != nil {
			log.Debugf("failed to parse mac address on interface: %s, mac: %s", m.Interface, m.MACAddress)
			continue
		}
		macTable[m.MACAddress] = &MACTableEntry{
			Ifname: m.Interface,
			VLAN:   strconv.Itoa(m.VlanID),
			MAC:    mac,
			Type:   m.EntryType,
			Port:   port,
		}
	}

	return macTable, nil
}

type ShowLLDPNeighborsDetail struct {
	LLDPNeighbors map[string]struct {
		LLDPNeighborInfo []LLDPNeighborInfo `json:"lldpNeighborInfo"`
	} `json:"lldpNeighbors"`
}

type LLDPNeighborInfo struct {
	ChassisIdType      string `json:"chassisIdType"`
	ChassisId          string `json:"chassisId"`
	SystemName         string `json:"systemName"`
	SystemDescription  string `json:"systemDescription"`
	SystemCapabilities struct {
		Bridge bool `json:"bridge"`
		Router bool `json:"router"`
	} `json:"systemCapabilities"`
	LastContactTime       float32 `json:"lastContactTime"`
	NeighborDiscoveryTime float32 `json:"neighborDiscoveryTime"`
	LastChangeTime        float32 `json:"lastChangeTime"`
	TTL                   int     `json:"ttl"`
	ManagementAddresses   []struct {
		AddressType           string `json:"addressType"`
		Address               string `json:"address"`
		InterfaceNumType      string `json:"interfaceNumType"`
		InterfaceNumTypeValue int    `json:"interfaceNumTypeValue"`
		InterfaceNum          int    `json:"interfaceNum"`
		OidString             string `json:"oidString"`
	} `json:"managementAddress"`
	NeighborInterfaceInfo struct {
		InterfaceIdType      string `json:"interfaceIdType"`
		InterfaceId          string `json:"interfaceId"`
		InterfaceId_v2       string `json:"interfaceId_v2"`
		InterfaceDescription string `json:"interfaceDescription"`
		MaxFrameSize         int    `json:"maxFrameSize"`
		PortVlanId           int    `json:"portVlanId"`
	} `json:"neighborInterfaceInfo"`
}

func (s *ShowLLDPNeighborsDetail) GetCmd() string {
	return "show lldp neighbors detail"
}

func (a *Arista) GetLLDPNeighbors() (LLDPNeighbors, error) {
	cmd := &ShowLLDPNeighborsDetail{}

	handle, err := a.client.GetHandle("json")
	if err != nil {
		return nil, err
	}
	err = handle.AddCommand(cmd)
	if err != nil {
		return nil, err
	}
	err = handle.Call()
	if err != nil {
		return nil, err
	}

	lldp := make(LLDPNeighbors, 0)

	for key, iface := range cmd.LLDPNeighbors {
		// name := strings.Replace(key, "Ethernet", "", 1)
		for _, neighborInfo := range iface.LLDPNeighborInfo {
			mgmtAddress := ""
			if len(neighborInfo.ManagementAddresses) == 1 {
				mgmtAddress = neighborInfo.ManagementAddresses[0].Address
			}
			mac, err := net.ParseMAC(neighborInfo.ChassisId)
			if err != nil {
				log.Errorf("Failed to parse mac from switch lldp query. Switch:%s error:%s", key, err)
				continue
			}
			lldp[key] = &LLDP{
				ChassisId:         mac,
				ChassisIdType:     neighborInfo.ChassisIdType,
				SystemName:        neighborInfo.SystemName,
				SystemDescription: neighborInfo.SystemDescription,
				ManagementAddress: mgmtAddress,
				PortDescription:   neighborInfo.NeighborInterfaceInfo.InterfaceDescription,
				PortId:            neighborInfo.NeighborInterfaceInfo.InterfaceId_v2,
				PortIdType:        neighborInfo.NeighborInterfaceInfo.InterfaceIdType,
			}
		}
	}
	return lldp, nil
}
