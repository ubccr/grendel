package model

import (
	"io/ioutil"
)

type Datastore interface {
	DefaultBootSpec() (*BootSpec, error)
	GetBootSpec(mac string) (*BootSpec, error)
}

type StaticBooter struct {
	bootSpec *BootSpec
}

func (s *StaticBooter) DefaultBootSpec() (*BootSpec, error) {
	return s.bootSpec, nil
}

func (s *StaticBooter) GetBootSpec(mac string) (*BootSpec, error) {
	return s.bootSpec, nil
}

func NewStaticBooter(kernelPath string, initrdPaths []string, cmdline string) (*StaticBooter, error) {
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

	booter := &StaticBooter{bootSpec: spec}

	return booter, nil
}
