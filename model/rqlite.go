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

package model

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/rqlite/gorqlite"
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/util"
)

// RQLite implements a Grendel Datastore using BuntDB
type RQLite struct {
	db *gorqlite.Connection
}

// NewRQLite returns a new RQLite using the given database filename. For memory only you can provide `:memory:`
func NewRqliteStore(addr string) (*RQLite, error) {
	db, err := gorqlite.Open(addr)
	if err != nil {
		return nil, err
	}

	// Check for DB migration
	err = migrate(db)
	if err != nil {
		log.Debugf("rqlite migration failed: %s", err)
		return nil, err

	}

	// db.SetConsistencyLevel("strong")

	return &RQLite{db: db}, nil
}

func migrate(db *gorqlite.Connection) error {
	version, err := getVersion(db)
	if err != nil {
		return err
	}

	switch version {
	case 0:
		// New DB init
		log.Debugf("rqlite user_version is 0! Initializing DB tables")

		statements := []string{
			"CREATE TABLE IF NOT EXISTS users (username TEXT PRIMARY KEY, hash TEXT, role TEXT, created_at DATETIME, modified_at DATETIME)",
		}

		_, err := db.Write(statements)
		if err != nil {
			return err
		}

		err = setVersion(db, 2)
		if err != nil {
			return err
		}
		version = 2
		log.Debugf("rqlite successfully updated tables to version 2.")
	case 1:
		// Test table datatype migration (worse case)
		log.Debugf("rqlite updating tables from version 1 to 2.")

		statements := []string{
			"CREATE TABLE IF NOT EXISTS users_migration (username TEXT PRIMARY KEY, hash TEXT, role TEXT, created_at TEXT, modified_at TEXT)",
			"INSERT INTO users_migration SELECT * FROM users",
			"DROP TABLE users",
			"CREATE TABLE users (username TEXT PRIMARY KEY, hash TEXT, role TEXT, created_at DATETIME, modified_at DATETIME)",
			"INSERT INTO users SELECT * FROM users_migration",
			"DROP TABLE users_migration",
		}

		_, err := db.Write(statements)
		if err != nil {
			return err
		}

		err = setVersion(db, 2)
		if err != nil {
			return err
		}
		version += 1
		log.Debugf("rqlite successfully updated tables to version 2.")
		fallthrough
	case 2:
		log.Debug("rqlite tables are up to date!")
	default:
		return errors.New("rqlite unknown database version")
	}

	return nil
}

func getVersion(db *gorqlite.Connection) (int, error) {
	var version int

	qr, err := db.QueryOne("PRAGMA user_version")
	if err != nil {
		return 0, err
	}
	for qr.Next() {
		err := qr.Scan(&version)
		if err != nil {
			return 0, err
		}
	}

	log.Debugf("rqlite database version: %d", version)
	return version, nil
}

func setVersion(db *gorqlite.Connection, version int32) error {
	// TODO: Parameterized write does not work?
	_, err := db.WriteOne(fmt.Sprintf("PRAGMA user_version = %d", version))
	if err != nil {
		return err
	}

	return nil
}

// Close closes the RQLite database
func (s *RQLite) Close() error {
	s.db.Close()

	return nil
}

// StoreUser stores the User in the data store
func (s *RQLite) StoreUser(username, password string) error {
	role := "disabled"

	// Set role to admin if this is the first user
	qr, err := s.db.QueryOne("SELECT username FROM users")

	if err != nil {
		return err
	}
	if qr.NumRows() == 0 {
		role = "admin"
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 8)
	_, err = s.db.WriteOneParameterized(
		gorqlite.ParameterizedStatement{
			Query: "INSERT INTO users(username, hash, role, created_at, modified_at) VALUES(?, ?, ?, ?, ?)",
			Arguments: []interface{}{
				username,
				string(hashed),
				role,
				time.Now(),
				time.Now(),
			},
		},
	)

	log.Debugf("rqlite.StoreUser: inserted '%s' user", username)
	return err
}

// VerifyUser checks if the given username exists in the data store
func (s *RQLite) VerifyUser(username, password string) (bool, string, error) {
	var hash string
	var role string

	qr, err := s.db.QueryOneParameterized(
		gorqlite.ParameterizedStatement{
			Query: "SELECT hash, role FROM users WHERE username = ?",
			Arguments: []interface{}{
				username,
			},
		},
	)

	if err != nil {
		return false, "", err
	}

	for qr.Next() {
		err := qr.Scan(&hash, &role)
		if err != nil {
			return false, "", err
		}
	}

	log.Debugf("rqlite.VerifyUser: queried %d user(s)", qr.NumRows())

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return true, role, err
}

// GetUsers returns a list of all the usernames
func (s *RQLite) GetUsers() ([]User, error) {
	var users []User

	qr, err := s.db.QueryOne("SELECT username, role, created_at, modified_at FROM users")
	if err != nil {
		return nil, err
	}

	for qr.Next() {
		var u User
		err := qr.Scan(&u.Username, &u.Role, &u.CreatedAt, &u.ModifiedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	log.Debugf("rqlite.GetUsers: queried %d user(s)", qr.NumRows())
	return users, nil
}

// UpdateUser updates the role of the given users
func (s *RQLite) UpdateUser(username, role string) error {
	wr, err := s.db.WriteOneParameterized(gorqlite.ParameterizedStatement{
		Query: "UPDATE users SET role = ? WHERE username = ?",
		Arguments: []interface{}{
			role,
			username,
		},
	})
	log.Debugf("rqlite.UpdateUser: updated %d user(s) role", wr.RowsAffected)
	return err
}

// DeleteUser deletes the given user
func (s *RQLite) DeleteUser(username string) error {
	_, err := s.db.WriteOneParameterized(gorqlite.ParameterizedStatement{
		Query: "DELETE FROM users WHERE username = ?",
		Arguments: []interface{}{
			username,
		},
	})
	log.Debugf("rqlite.DeleteUser: deleted %s user", username)
	return err
}

// StoreHost stores a host in the data store. If the host exists it is overwritten
func (s *RQLite) StoreHost(host *Host) error {
	hostList := HostList{host}
	return s.StoreHosts(hostList)
}

// StoreHosts stores a list of host in the data store. If the host exists it is overwritten
func (s *RQLite) StoreHosts(hosts HostList) error {

	return nil
}

// DeleteHosts deletes all hosts in the given nodeset.NodeSet from the data store.
func (s *RQLite) DeleteHosts(ns *nodeset.NodeSet) error {

	return nil
}

// LoadHostFromName returns the Host with the given name
func (s *RQLite) LoadHostFromName(name string) (*Host, error) {
	var host *Host

	return host, nil
}

// LoadHostFromID returns the Host with the given ID
func (s *RQLite) LoadHostFromID(id string) (*Host, error) {
	hostJSON := ""
	host := &Host{}
	host.FromJSON(hostJSON)
	return host, nil
}

// ResolveIPv4 returns the list of IPv4 addresses with the given FQDN
func (s *RQLite) ResolveIPv4(fqdn string) ([]net.IP, error) {
	fqdn = util.Normalize(fqdn)
	ips := make([]net.IP, 0)

	return ips, nil
}

// ReverseResolve returns the list of FQDNs for the given IP
func (s *RQLite) ReverseResolve(ip string) ([]string, error) {
	fqdn := make([]string, 0)

	return fqdn, nil
}

// LoadHostFromMAC returns the Host that has a network interface with the give MAC address
func (s *RQLite) LoadHostFromMAC(mac string) (*Host, error) {
	hostJSON := ""

	host := &Host{}
	host.FromJSON(hostJSON)
	return host, nil
}

// Hosts returns a list of all the hosts
func (s *RQLite) Hosts() (HostList, error) {
	hosts := make(HostList, 0)

	return hosts, nil
}

// FindHosts returns a list of all the hosts in the given NodeSet
func (s *RQLite) FindHosts(ns *nodeset.NodeSet) (HostList, error) {
	hosts := make(HostList, 0)

	return hosts, nil
}

// FindTags returns a nodeset.NodeSet of all the hosts with the given tags
func (s *RQLite) FindTags(tags []string) (*nodeset.NodeSet, error) {
	nodes := []string{}

	tagMap := make(map[string]struct{})
	for _, t := range tags {
		tagMap[t] = struct{}{}
	}

	return nodeset.NewNodeSet(strings.Join(nodes, ","))
}

// MatchTags returns a nodeset.NodeSet of all the hosts with the all given tags
func (s *RQLite) MatchTags(tags []string) (*nodeset.NodeSet, error) {
	nodes := []string{}

	return nodeset.NewNodeSet(strings.Join(nodes, ","))
}

// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
func (s *RQLite) ProvisionHosts(ns *nodeset.NodeSet, provision bool) error {

	return nil
}

// TagHosts adds tags to all hosts in the given NodeSet
func (s *RQLite) TagHosts(ns *nodeset.NodeSet, tags []string) error {
	return nil
}

// UntagHosts removes tags from all hosts in the given NodeSet
func (s *RQLite) UntagHosts(ns *nodeset.NodeSet, tags []string) error {
	return nil
}

// SetBootImage sets all hosts to use the BootImage with the given name
func (s *RQLite) SetBootImage(ns *nodeset.NodeSet, name string) error {
	return nil
}

// StoreBootImage stores a boot image in the data store. If the boot image exists it is overwritten
func (s *RQLite) StoreBootImage(image *BootImage) error {
	imageList := BootImageList{image}
	return s.StoreBootImages(imageList)
}

// StoreBootImages stores a list of boot images in the data store. If the boot image exists it is overwritten
func (s *RQLite) StoreBootImages(images BootImageList) error {
	return nil
}

// DeleteBootImages deletes boot images from the data store.
func (s *RQLite) DeleteBootImages(names []string) error {
	return nil
}

// LoadBootImage returns a BootImage with the given name
func (s *RQLite) LoadBootImage(name string) (*BootImage, error) {
	var image *BootImage

	return image, nil
}

// BootImages returns a list of all boot images
func (s *RQLite) BootImages() (BootImageList, error) {
	images := make(BootImageList, 0)

	return images, nil
}
