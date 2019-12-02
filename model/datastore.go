package model

import (
	"github.com/ubccr/grendel/logger"
)

var log = logger.GetLogger("DB")

type Datastore interface {
	DefaultBootSpec() (*BootSpec, error)
	GetBootSpec(mac string) (*BootSpec, error)
	GetHost(mac string) (*Host, error)
}
