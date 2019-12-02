package model

import (
	"fmt"
	"io/ioutil"
)

type BootSpec struct {
	Name    string
	Kernel  []byte
	Initrd  [][]byte
	Cmdline string
	Message string
}

type StaticBooter struct {
	bootSpec *BootSpec
	hostList map[string]*Host
}

func (s *StaticBooter) DefaultBootSpec() (*BootSpec, error) {
	return s.bootSpec, nil
}

func (s *StaticBooter) GetBootSpec(mac string) (*BootSpec, error) {
	return s.bootSpec, nil
}

func NewStaticBooter(filename, kernelPath string, initrdPaths []string, cmdline string) (*StaticBooter, error) {
	kernel, err := ioutil.ReadFile(kernelPath)
	if err != nil {
		return nil, err
	}

	initrds := make([][]byte, 0)
	for _, img := range initrdPaths {
		data, err := ioutil.ReadFile(img)
		if err != nil {
			return nil, err
		}
		initrds = append(initrds, data)
	}

	spec := &BootSpec{
		Kernel:  kernel,
		Initrd:  initrds,
		Cmdline: cmdline,
	}

	hostList, err := ParseStaticHostList(filename)
	if err != nil {
		return nil, err
	}

	booter := &StaticBooter{bootSpec: spec, hostList: hostList}

	return booter, nil
}

func (s *StaticBooter) GetHost(mac string) (*Host, error) {
	if host, ok := s.hostList[mac]; ok {
		return host, nil
	}

	return nil, fmt.Errorf("Host not found with hwaddr: %s", mac)
}
