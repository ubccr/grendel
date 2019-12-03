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

func NewStaticBooter(filename, kernelPath string, initrdPaths []string, cmdline, liveImage string) (*StaticBooter, error) {
	image := &BootImage{
		KernelPath:  kernelPath,
		InitrdPaths: initrdPaths,
		CommandLine: cmdline,
		LiveImage:   liveImage,
	}

	hostList, err := ParseStaticHostList(filename)
	if err != nil {
		return nil, err
	}

	booter := &StaticBooter{bootImage: image, hostList: hostList}

	return booter, nil
}

func (s *StaticBooter) GetHost(mac string) (*Host, error) {
	if host, ok := s.hostList[mac]; ok {
		return host, nil
	}

	return nil, fmt.Errorf("Host not found with hwaddr: %s", mac)
}
