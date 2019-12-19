package model

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/ubccr/go-dhcpd-leases"
)

type HostList []*Host

func (hl HostList) FilterPrefix(prefix string) HostList {
	n := 0
	for _, host := range hl {
		if strings.HasPrefix(host.Name, prefix) {
			hl[n] = host
			n++
		}
	}

	return hl[:n]
}

func ParseStaticHostList(filename string) (map[string]*Host, error) {
	hostList := make(map[string]*Host)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")
		hwaddr, err := net.ParseMAC(cols[0])
		if err != nil {
			return nil, fmt.Errorf("Malformed hardware address: %s", cols[0])
		}
		ipaddr := net.ParseIP(cols[1])
		if ipaddr.To4() == nil {
			return nil, fmt.Errorf("Invalid IPv4 address: %v", cols[1])
		}

		nic := &NetInterface{MAC: hwaddr, IP: ipaddr}

		if len(cols) > 2 {
			nic.FQDN = cols[2]
		}

		uuid, err := ksuid.NewRandom()
		if err != nil {
			return nil, err
		}

		host := &Host{ID: uuid}
		host.Interfaces = []*NetInterface{nic}

		if len(cols) > 3 && strings.ToLower(cols[3]) == "yes" {
			host.Provision = true
		}

		hostList[host.ID.String()] = host
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hostList, nil
}

func ParseLeases(filename string) (map[string]*Host, error) {
	hostList := make(map[string]*Host)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hosts := leases.Parse(file)
	if hosts == nil {
		return nil, errors.New("No hosts found. Is this a dhcpd.leasts file?")
	}

	for _, h := range hosts {
		nic := &NetInterface{MAC: h.Hardware.MACAddr, IP: h.IP}

		if len(h.ClientHostname) > 0 {
			nic.FQDN = h.ClientHostname
		}

		uuid, err := ksuid.NewRandom()
		if err != nil {
			return nil, err
		}

		host := &Host{ID: uuid}
		host.Interfaces = []*NetInterface{nic}

		hostList[host.ID.String()] = host
	}

	return hostList, nil
}
