// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

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

// DataStore
type DataStore interface {
	// BootImages returns a list of all boot images
	BootImages() (BootImageList, error)

	// LoadBootImage returns a BootImage with the given name
	LoadBootImage(name string) (*BootImage, error)

	// StoreBootImage stores the BootImage in the data store
	StoreBootImage(image *BootImage) error

	// StoreBootImages stores a list of BootImages in the data store
	StoreBootImages(images BootImageList) error

	// SetBootImage sets all hosts to use the BootImage with the given name
	SetBootImage(ns *nodeset.NodeSet, name string) error

	// Hosts returns a list of all the hosts
	Hosts() (HostList, error)

	// FindHosts returns a list of all the hosts in the given NodeSet
	FindHosts(ns *nodeset.NodeSet) (HostList, error)

	// FindTags returns a nodeset.NodeSet of all the hosts with the given tags
	FindTags(tags []string) (*nodeset.NodeSet, error)

	// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
	ProvisionHosts(ns *nodeset.NodeSet, provision bool) error

	// TagHosts adds tags to all hosts in the given NodeSet
	TagHosts(ns *nodeset.NodeSet, tags []string) error

	// UntagHosts removes tags from all hosts in the given NodeSet
	UntagHosts(ns *nodeset.NodeSet, tags []string) error

	// StoreHosts stores a hosts in the data store. If the host exists it is overwritten
	StoreHost(host *Host) error

	// StoreHosts stores a list of hosts in the data store. If the host exists it is overwritten
	StoreHosts(hosts HostList) error

	// LoadHostFromID returns the Host with the given ID
	LoadHostFromID(id string) (*Host, error)

	// LoadHostFromName returns the Host with the given name
	LoadHostFromName(name string) (*Host, error)

	// LoadHostFromMAC returns the Host that has a network interface with the give MAC address
	LoadHostFromMAC(mac string) (*Host, error)

	// ResolveIPv4 returns the list of IPv4 addresses with the given FQDN
	ResolveIPv4(fqdn string) ([]net.IP, error)

	// ReverseResolve returns the list of FQDNs for the given IP
	ReverseResolve(ip string) ([]string, error)

	// Close data store
	Close() error
}

func NewDataStore(path string) (DataStore, error) {
	return NewBuntStore(path)
}
