package model

import (
	"fmt"
	"sync"

	"github.com/segmentio/ksuid"
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

func (s *StaticBooter) LoadStaticHosts(filename string) error {
	hostList, err := ParseStaticHostList(filename)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

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

	s.Lock()
	defer s.Unlock()

	for k, v := range hostList {
		s.hostList[k] = v
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
