package tors

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

func (d *Sonic) GetMACTable() (MACTable, error) {
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

	macTable := make(MACTable, 0)

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

		macTable[entry.MacAddress] = &MACTableEntry{
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
		iface []struct {
			name      string
			neighbors []struct {
				neighbor struct {
					id    string
					state struct {
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
					}
				}
			}
		}
	} `json:"openconfig-lldp:interfaces"`
}

// TODO:
// func (s *Sonic) GetLLDPNeighbors() (LLDPNeighbors, error) {
// 	var lldpRaw *sonicLLDP
// 	res, err := s.getRequest(SONIC_RESTCONF_LLDP)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = json.Unmarshal(res, &lldpRaw)
// 	if err != nil {
// 		return nil, err
// 	}

// 	o := make(LLDPNeighbors, 0)

// 	for _, iface := range lldpRaw.Root.iface {
// 		for _, n := range iface.neighbors {
// 			state := n.neighbor.state
// 			o[iface.name] = &LLDP{
// 				ChassisId:         state.ChassisId,
// 				ChassisIdType:     state.ChassisidType,
// 				ManagementAddress: state.ManagementAddress,
// 				PortDescription:   state.PortDescription,
// 				PortId:            state.PortId,
// 				PortIdType:        state.PortIdType,
// 				SystemDescription: state.SystemDescription,
// 				SystemName:        state.SystemName,
// 			}
// 		}
// 	}

// 	return nil, nil
// }
