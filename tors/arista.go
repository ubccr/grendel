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
