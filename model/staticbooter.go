package model

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"sync"

	"github.com/segmentio/ksuid"
	"github.com/ubccr/go-dhcpd-leases"
)

type StaticBooter struct {
	sync.RWMutex

	bootImage *BootImage
	hostList  map[string]*Host
}

func (s *StaticBooter) GetBootImage(mac string) (*BootImage, error) {
	return s.bootImage, nil
}

func NewStaticBooter(kernelPath string, initrdPaths []string, cmdline, liveImage, rootFS, installRepo string) (*StaticBooter, error) {
	image := &BootImage{
		KernelPath:  kernelPath,
		InitrdPaths: initrdPaths,
		CommandLine: cmdline,
		LiveImage:   liveImage,
		RootFS:      rootFS,
		InstallRepo: installRepo,
	}

	booter := &StaticBooter{bootImage: image, hostList: make(map[string]*Host)}

	return booter, nil
}

func (s *StaticBooter) LoadStaticHosts(reader io.Reader) error {
	s.Lock()
	defer s.Unlock()

	s.hostList = make(map[string]*Host)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")
		hwaddr, err := net.ParseMAC(cols[0])
		if err != nil {
			return fmt.Errorf("Malformed hardware address: %s", cols[0])
		}
		ipaddr := net.ParseIP(cols[1])
		if ipaddr.To4() == nil {
			return fmt.Errorf("Invalid IPv4 address: %v", cols[1])
		}

		nic := &NetInterface{MAC: hwaddr, IP: ipaddr}

		if len(cols) > 2 {
			nic.FQDN = cols[2]
		}

		uuid, err := ksuid.NewRandom()
		if err != nil {
			return err
		}

		host := &Host{ID: uuid}
		host.Interfaces = []*NetInterface{nic}

		if len(cols) > 3 && strings.ToLower(cols[3]) == "yes" {
			host.Provision = true
		}

		s.hostList[host.ID.String()] = host
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (s *StaticBooter) LoadDHCPLeases(reader io.Reader) error {
	hosts := leases.Parse(reader)
	if hosts == nil {
		return errors.New("No hosts found. Is this a dhcpd.leasts file?")
	}

	s.Lock()
	defer s.Unlock()

	s.hostList = make(map[string]*Host)
	for _, h := range hosts {
		nic := &NetInterface{MAC: h.Hardware.MACAddr, IP: h.IP}

		if len(h.ClientHostname) > 0 {
			nic.FQDN = h.ClientHostname
		}

		uuid, err := ksuid.NewRandom()
		if err != nil {
			return err
		}

		host := &Host{ID: uuid}
		host.Interfaces = []*NetInterface{nic}

		s.hostList[host.ID.String()] = host
	}

	return nil
}

func (s *StaticBooter) LoadJSON(reader io.Reader) error {
	jsonBlob, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var hostList HostList
	err = json.Unmarshal(jsonBlob, &hostList)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	s.hostList = make(map[string]*Host)
	for _, h := range hostList {
		if h.ID.IsNil() {
			uuid, err := ksuid.NewRandom()
			if err != nil {
				return err
			}

			h.ID = uuid
		}
		s.hostList[h.ID.String()] = h
	}

	return nil
}

func (s *StaticBooter) GetHost(id string) (*Host, error) {
	s.RLock()
	defer s.RUnlock()

	if host, ok := s.hostList[id]; ok {
		return host, nil
	}

	return nil, fmt.Errorf("Host not found with id: %s", id)
}

func (s *StaticBooter) SaveHost(host *Host) error {
	s.Lock()
	defer s.Unlock()

	if s.hostList == nil {
		s.hostList = make(map[string]*Host)
	}

	if host.ID.IsNil() {
		uuid, err := ksuid.NewRandom()
		if err != nil {
			return err
		}
		host.ID = uuid
	}

	s.hostList[host.ID.String()] = host

	return nil
}

func (s *StaticBooter) HostList() (HostList, error) {
	s.RLock()
	defer s.RUnlock()

	values := make(HostList, 0, len(s.hostList))

	for _, v := range s.hostList {
		values = append(values, v)
	}
	return values, nil
}
