// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package model provides the data model for grendel
package store

import (
	"net"

	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

var (
	// Global logger for Store package
	Log = logger.GetLogger("STORE")
)

type Store interface {
	// StoreUser stores the User in the data store
	StoreUser(username, password string) (string, error)

	// VerifyUser checks if the given username exists in the data store
	VerifyUser(username, password string) (bool, string, error)

	// GetUsers returns a list of all the usernames
	GetUsers() ([]model.User, error)

	// GetUserByName returns a user
	GetUserByName(name string) (*model.User, error)

	// UpdateUserRole updates the role of the given users
	UpdateUserRole(username, role string) error

	// UpdateUserEnabled updates the role of the given users
	UpdateUserEnabled(username string, enabled bool) error

	// DeleteUser deletes the given user
	DeleteUser(username string) error

	// BootImages returns a list of all boot images
	BootImages() (model.BootImageList, error)

	// LoadBootImage returns a BootImage with the given name
	LoadBootImage(name string) (*model.BootImage, error)

	// StoreBootImage stores the BootImage in the data store
	StoreBootImage(image *model.BootImage) error

	// StoreBootImages stores a list of BootImages in the data store
	StoreBootImages(images model.BootImageList) error

	// DeleteBootImages delete BootImages from the data store
	DeleteBootImages(names []string) error

	// SetBootImage sets all hosts to use the BootImage with the given name
	SetBootImage(ns *nodeset.NodeSet, name string) error

	// Hosts returns a list of all the hosts
	Hosts() (model.HostList, error)

	// FindHosts returns a list of all the hosts in the given NodeSet
	FindHosts(ns *nodeset.NodeSet) (model.HostList, error)

	// FindTags returns a nodeset.NodeSet of all the hosts with the given tags
	FindTags(tags []string) (*nodeset.NodeSet, error)

	// MatchTags returns a nodeset.NodeSet of all the hosts with the all given tags
	MatchTags(tags []string) (*nodeset.NodeSet, error)

	// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
	ProvisionHosts(ns *nodeset.NodeSet, provision bool) error

	// TagHosts adds tags to all hosts in the given NodeSet
	TagHosts(ns *nodeset.NodeSet, tags []string) error

	// UntagHosts removes tags from all hosts in the given NodeSet
	UntagHosts(ns *nodeset.NodeSet, tags []string) error

	// StoreHosts stores a host in the data store. If the host exists it is overwritten
	StoreHost(host *model.Host) error

	// StoreHosts stores a list of hosts in the data store. If the host exists it is overwritten
	StoreHosts(hosts model.HostList) error

	// DeleteHosts deletes all hosts in the given nodeset.NodeSet from the data store.
	DeleteHosts(ns *nodeset.NodeSet) error

	// LoadHostFromID returns the Host with the given ID
	LoadHostFromID(id string) (*model.Host, error)

	// LoadHostFromName returns the Host with the given name
	LoadHostFromName(name string) (*model.Host, error)

	// LoadHostFromMAC returns the Host that has a network interface with the give MAC address
	LoadHostFromMAC(mac string) (*model.Host, error)

	// ResolveIPv4 returns the list of IPv4 addresses with the given FQDN
	ResolveIPv4(fqdn string) ([]net.IP, error)

	// ReverseResolve returns the list of FQDNs for the given IP
	ReverseResolve(ip string) ([]string, error)

	// RestoreFrom restores the database using the provided data dump
	RestoreFrom(data model.DataDump) error

	// GetRolesByRoute returns all roles that have access to the provided method and path
	GetRolesByRoute(method, path string) (*[]string, error)

	// GetRoles returns all roles
	GetRoles() (model.RoleViewList, error)

	// GetRoleByName returns a role by name
	GetRolesByName(name string) (*model.RoleView, error)

	// GetPermissions returns all permissions
	GetPermissions() (model.PermissionList, error)

	// AddRole adds a new role and optionally sets the same permissions as the inherited role
	AddRole(role, inheritedRole string) error

	// DeleteRole removes a role by name
	DeleteRole(roles []string) error

	// UpdateRolePermissions sets the permissions for the given role
	UpdateRolePermissions(role string, permissions model.PermissionList) error

	Close() error
}
