package model

import (
	"fmt"
)

type StaticBooter struct {
	bootImage *BootImage
	hostList  map[string]*Host
}

func (s *StaticBooter) GetBootImage(mac string) (*BootImage, error) {
	return s.bootImage, nil
}

func NewStaticBooter(kernelPath string, initrdPaths []string, cmdline, liveImage string) (*StaticBooter, error) {
	image := &BootImage{
		KernelPath:  kernelPath,
		InitrdPaths: initrdPaths,
		CommandLine: cmdline,
		LiveImage:   liveImage,
	}

	booter := &StaticBooter{bootImage: image, hostList: make(map[string]*Host)}

	return booter, nil
}

func (s *StaticBooter) LoadStaticHosts(filename string) error {
	hostList, err := ParseStaticHostList(filename)
	if err != nil {
		return err
	}

	for k, v := range hostList {
		s.hostList[k] = v
	}

	return nil
}

func (s *StaticBooter) LoadDHCPLeases(filename string) error {
	hostList, err := ParseLeases(filename)
	if err != nil {
		return err
	}

	for k, v := range hostList {
		s.hostList[k] = v
	}

	return nil
}

func (s *StaticBooter) GetHost(mac string) (*Host, error) {
	if host, ok := s.hostList[mac]; ok {
		return host, nil
	}

	return nil, fmt.Errorf("Host not found with hwaddr: %s", mac)
}

func (s *StaticBooter) SaveHost(host *Host) error {
	if s.hostList == nil {
		s.hostList = make(map[string]*Host)
	}

	s.hostList[host.MAC.String()] = host
	return nil
}

func (s *StaticBooter) HostList() ([]*Host, error) {
	values := make([]*Host, 0, len(s.hostList))

	for _, v := range s.hostList {
		values = append(values, v)
	}
	return values, nil
}
