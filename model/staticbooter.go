// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

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
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/util"
)

type StaticBooter struct {
	DefaultBooter string
	bootImageMap  sync.Map
	hostMap       sync.Map
}

func (s *StaticBooter) LoadHostTSV(reader io.Reader) error {
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

		s.hostMap.Store(host.Name, host)
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

		s.hostMap.Store(host.Name, host)
	}

	return nil
}

func (s *StaticBooter) LoadHostJSON(reader io.Reader) error {
	jsonBlob, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var hostList HostList
	err = json.Unmarshal(jsonBlob, &hostList)
	if err != nil {
		return err
	}

	for _, h := range hostList {
		if h.ID.IsNil() {
			uuid, err := ksuid.NewRandom()
			if err != nil {
				return err
			}

			h.ID = uuid
		}
		s.hostMap.Store(h.Name, h)
	}

	return nil
}

func (s *StaticBooter) LoadBootImageJSON(reader io.Reader) error {
	jsonBlob, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var bootImageList BootImageList
	err = json.Unmarshal(jsonBlob, &bootImageList)
	if err != nil {
		return err
	}

	for _, bi := range bootImageList {
		if bi.ID.IsNil() {
			uuid, err := ksuid.NewRandom()
			if err != nil {
				return err
			}

			bi.ID = uuid
		}
		s.bootImageMap.Store(bi.Name, bi)
	}

	return nil
}

func (s *StaticBooter) LoadHostFromName(name string) (*Host, error) {
	if host, ok := s.hostMap.Load(name); ok {
		return host.(*Host), nil
	}

	return nil, ErrNotFound
}

func (s *StaticBooter) LoadHostFromID(id string) (*Host, error) {
	var host *Host

	s.hostMap.Range(func(key, value interface{}) bool {
		h := value.(*Host)
		if id == h.ID.String() {
			host = h
			return false
		}

		return true
	})

	if host == nil {
		return nil, ErrNotFound
	}

	return host, nil
}

func (s *StaticBooter) StoreHosts(hosts HostList) error {
	for _, host := range hosts {
		err := s.StoreHost(host)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StaticBooter) StoreHost(host *Host) error {
	if host.ID.IsNil() {
		uuid, err := ksuid.NewRandom()
		if err != nil {
			return err
		}
		host.ID = uuid
	}

	s.hostMap.Store(host.Name, host)

	return nil
}

func (s *StaticBooter) Hosts() (HostList, error) {
	values := make(HostList, 0)

	s.hostMap.Range(func(key, value interface{}) bool {
		values = append(values, value.(*Host))
		return true
	})

	return values, nil
}

func (s *StaticBooter) FindHosts(ns *nodeset.NodeSet) (HostList, error) {
	values := make(HostList, 0)

	it := ns.Iterator()
	for it.Next() {
		host, err := s.LoadHostFromName(it.Value())
		if err == nil {
			values = append(values, host)
		}
	}

	return values, nil
}

func (s *StaticBooter) BootImages() (BootImageList, error) {
	values := make(BootImageList, 0)

	s.bootImageMap.Range(func(key, value interface{}) bool {
		values = append(values, value.(*BootImage))
		return true
	})

	return values, nil
}

func (s *StaticBooter) LoadBootImage(name string) (*BootImage, error) {
	if name == "" {
		name = s.DefaultBooter
	}

	if bootImage, ok := s.bootImageMap.Load(name); ok {
		return bootImage.(*BootImage), nil
	}

	return nil, ErrNotFound
}

func (s *StaticBooter) DefaultBootImage() (*BootImage, error) {
	return s.LoadBootImage(s.DefaultBooter)
}

func (s *StaticBooter) StoreBootImage(bootImage *BootImage) error {
	if bootImage.ID.IsNil() {
		uuid, err := ksuid.NewRandom()
		if err != nil {
			return err
		}
		bootImage.ID = uuid
	}

	s.bootImageMap.Store(bootImage.Name, bootImage)

	return nil
}

func (s *StaticBooter) SetBootImage(ns *nodeset.NodeSet, imageName string) error {
	image, err := s.LoadBootImage(imageName)
	if err != nil {
		return err
	}

	it := ns.Iterator()
	for it.Next() {
		host, err := s.LoadHostFromName(it.Value())
		if err == nil {
			host.BootImage = image.Name
		}
	}

	return nil
}

func (s *StaticBooter) ProvisionHosts(ns *nodeset.NodeSet, provision bool) error {
	it := ns.Iterator()
	for it.Next() {
		host, err := s.LoadHostFromName(it.Value())
		if err == nil {
			host.Provision = provision
		}
	}

	return nil
}

func (s *StaticBooter) LoadHostFromMAC(mac string) (*Host, error) {
	var host *Host

	s.hostMap.Range(func(key, value interface{}) bool {
		h := value.(*Host)
		for _, nic := range h.Interfaces {
			if nic.MAC.String() == mac {
				host = h
				return false
			}
		}
		return true
	})

	if host == nil {
		return nil, ErrNotFound
	}

	return host, nil
}

func (s *StaticBooter) ResolveIPv4(fqdn string) ([]net.IP, error) {
	fqdn = util.Normalize(fqdn)
	ips := make([]net.IP, 0)

	s.hostMap.Range(func(key, value interface{}) bool {
		h := value.(*Host)
		for _, nic := range h.Interfaces {
			if util.Normalize(nic.FQDN) == fqdn {
				ips = append(ips, nic.IP)
				return false
			}
		}
		return true
	})

	return ips, nil
}

func (s *StaticBooter) ReverseResolve(ip string) ([]string, error) {
	names := make([]string, 0)

	s.hostMap.Range(func(key, value interface{}) bool {
		h := value.(*Host)
		for _, nic := range h.Interfaces {
			if nic.IP.String() == ip {
				names = append(names, nic.FQDN)
				return false
			}
		}
		return true
	})

	return names, nil
}
