package model

import (
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/nodeset"
)

var log = logger.GetLogger("DB")

type Datastore interface {
	GetBootImage(mac string) (*BootImage, error)
	GetHostByID(id string) (*Host, error)
	GetHostByName(name string) (*Host, error)
	HostList() (HostList, error)
	Find(ns *nodeset.NodeSet) (HostList, error)
	SaveHost(host *Host) error
}
