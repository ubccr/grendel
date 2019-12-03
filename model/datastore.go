package model

import (
	"github.com/ubccr/grendel/logger"
)

var log = logger.GetLogger("DB")

type Datastore interface {
	GetBootImage(mac string) (*BootImage, error)
	GetHost(mac string) (*Host, error)
}
