// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package model provides the data model for grendel
package store

import (
	"context"
	"net"
	"time"

	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

var (
	// Global logger for Store package
	Log = logger.GetLogger("STORE")
)

type Store interface {
	// StoreUser creates a new user and returns its role. It fails with ErrDuplicateEntry when the username is taken, and never changes an existing account
	StoreUser(username, password string) (string, error)

	// UpdateUserPassword replaces the password of an existing user, failing with ErrNotFound when there is none
	UpdateUserPassword(username, password string) error

	// VerifyUser checks a username and password, giving the user's role when they match. An unknown user or a wrong password is not authenticated with no error; an error means the check itself failed
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

	// TagCounts returns every tag on a host, with how many hosts have it
	TagCounts() (model.TagCountList, error)

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
	ResolveIPv4(ctx context.Context, fqdn string) ([]net.IP, error)

	// ReverseResolve returns the list of FQDNs for the given IP
	ReverseResolve(ctx context.Context, ip string) ([]string, error)

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

	// LoadProvisionStatus returns the provision status of the host with the given ID, or ErrNotFound when it has none
	LoadProvisionStatus(uid string) (*model.ProvisionStatus, error)

	// UpdateProvisionStatus changes a host's provision status in one transaction. update is given the stored status, a zero one when the host has none, and an error from it is returned with nothing written. A node update returns is written in the same transaction to the running event of open's kind, which open is stored as when there is none, and that event is settled
	UpdateProvisionStatus(uid string, open *model.Event, update func(status *model.ProvisionStatus) (*model.EventNode, error)) error

	// ProvisionStatuses returns the provision status of every host in ns that has one, or of every host when ns is nil
	ProvisionStatuses(ns *nodeset.NodeSet) (model.ProvisionStatusList, error)

	// StoreEvent writes an event, replacing what can change about one already stored with the same ID, and writes the rows of its Nodes, leaving its other nodes as they were
	StoreEvent(event *model.Event) error

	// StoreEventNodes writes how an event is going on the given nodes, leaving its other nodes as they were, and settles the event in the same transaction. It is ErrNotFound when there is no such event
	StoreEventNodes(id string, nodes model.EventNodeList, now time.Time) error

	// SettleEvent brings an event up to date with how it is going on its nodes, as model.Event.Settle does, reading and writing in one transaction. It is ErrNotFound when there is no such event
	SettleEvent(id string, now time.Time) error

	// LoadEvent returns the event with the given ID with its nodes, or ErrNotFound
	LoadEvent(id string) (*model.Event, error)

	// Events returns the events the filter picks, newest first, with their counts, and with their nodes when the filter asks. It is ErrNotFound when the filter's Before names no event
	Events(filter model.EventFilter) (model.EventList, error)

	// StallEventNodes marks the nodes of running events of a kind as stalled when they were last updated before the given time, and returns how many
	StallEventNodes(kind model.EventKind, before time.Time) (int64, error)

	// PruneEvents deletes the finished events that started before the given time, and returns how many
	PruneEvents(before time.Time) (int64, error)

	Close() error

	// Settings returns the configuration stored in the database and a revision that changes whenever it does
	Settings() ([]settings.Row, int64, error)

	SettingsRevision() (int64, error)

	// InsertSettings stores the settings whose keys are not there yet, leaves the rest as they are, and returns the new revision
	InsertSettings(set []settings.Row, actor string) (int64, error)
	// UpdateSettings stores set and removes unset in one transaction, recording actor as who changed them, and returns the new revision
	UpdateSettings(set []settings.Row, unset []string, actor string) (int64, error)
}
