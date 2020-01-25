// Package model provides the data model for grendel
package model

import (
	"errors"
	"net"

	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/nodeset"
)

var (
	// Global logger for DB package
	log = logger.GetLogger("DB")

	// ErrNotFound is returned when a model is not found in the store
	ErrNotFound = errors.New("not found")

	// ErrInvalidData is returned when a model is is missing required data
	ErrInvalidData = errors.New("invalid data")
)

// Datastore
type Datastore interface {
	// BootImages returns a list of all boot images
	BootImages() (BootImageList, error)

	// LoadBootImage returns a BootImage with the given name
	LoadBootImage(name string) (*BootImage, error)

	// StoreBootImage stores the BootImage in the data store
	StoreBootImage(image *BootImage) error

	// SetBootImage sets all hosts to use the BootImage with the given name
	SetBootImage(ns *nodeset.NodeSet, name string) error

	// Hosts returns a list of all the hosts
	Hosts() (HostList, error)

	// FindHosts returns a list of all the hosts in the given NodeSet
	FindHosts(ns *nodeset.NodeSet) (HostList, error)

	// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
	ProvisionHosts(ns *nodeset.NodeSet, provision bool) error

	// StoreHost stores a host in the data store. If the host exists it is overwritten
	StoreHost(host *Host) error

	// LoadHostByID returns the Host with the given ID
	LoadHostByID(id string) (*Host, error)

	// LoadHostByName returns the Host with the given name
	LoadHostByName(name string) (*Host, error)

	// LoadHostByMAC returns the Host that has a network interface with the give MAC address
	LoadHostByMAC(mac string) (*Host, error)

	// LoadNetInterfaces returns the list of IPs with the given FQDN
	LoadNetInterfaces(fqdn string) ([]net.IP, error)
}
