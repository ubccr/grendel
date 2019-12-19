package model

import (
	"github.com/ubccr/grendel/logger"
)

var log = logger.GetLogger("DB")

type Datastore interface {
	GetBootImage(mac string) (*BootImage, error)
	GetHost(id string) (*Host, error)
	HostList() (HostList, error)
	SaveHost(host *Host) error
}
