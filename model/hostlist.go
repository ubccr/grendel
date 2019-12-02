package model

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

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

		host := &Host{MAC: hwaddr, IP: ipaddr}

		if len(cols) > 2 {
			host.FQDN = cols[2]
		}

		hostList[hwaddr.String()] = host
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hostList, nil
}
