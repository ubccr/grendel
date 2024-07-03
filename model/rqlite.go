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
	"net/netip"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/rqlite/gorqlite"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/util"
)

// RQLite implements a Grendel Datastore using BuntDB
type RQLite struct {
	db *gorqlite.Connection
}

// NewRQLite returns a new RQLite using the given database address.
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
			"CREATE TABLE IF NOT EXISTS hosts (name TEXT PRIMARY KEY, provision BOOL, firmware TEXT, boot_image TEXT)",
			"CREATE TABLE IF NOT EXISTS interfaces (id INT, host_name TEXT REFERENCES hosts (name) ON DELETE CASCADE, mac TEXT, name TEXT, ip TEXT, fqdn TEXT, bmc BOOL, vlan TEXT, mtu INT, PRIMARY KEY (id, host_name))",
			"CREATE TABLE IF NOT EXISTS bonds (id INT, host_name TEXT REFERENCES hosts (name) ON DELETE CASCADE, mac TEXT, name TEXT, ip TEXT, fqdn TEXT, bmc BOOL, vlan TEXT, mtu INT, peers TEXT, PRIMARY KEY (id, host_name))",
			"CREATE TABLE IF NOT EXISTS tags (host_name TEXT REFERENCES hosts(name) ON DELETE CASCADE, tag TEXT, value TEXT, PRIMARY KEY (host_name, tag))",
		}

		_, err := db.Write(statements)
		if err != nil {
			return err
		}

		err = setVersion(db, 3)
		if err != nil {
			return err
		}
		version = 3
		log.Debugf("rqlite successfully updated tables to version 3.")
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
		log.Debugf("rqlite updating tables from version 2 to 3.")

		statements := []string{
			"CREATE TABLE IF NOT EXISTS hosts (name TEXT PRIMARY KEY, provision BOOL, firmware TEXT, boot_image TEXT)",
			"CREATE TABLE IF NOT EXISTS interfaces (id INT, host_name TEXT REFERENCES hosts (name) ON DELETE CASCADE, mac TEXT, name TEXT, ip TEXT, fqdn TEXT, bmc BOOL, vlan TEXT, mtu INT, PRIMARY KEY (id, host_name))",
			"CREATE TABLE IF NOT EXISTS bonds (id INT, host_name TEXT REFERENCES hosts (name) ON DELETE CASCADE, mac TEXT, name TEXT, ip TEXT, fqdn TEXT, bmc BOOL, vlan TEXT, mtu INT, peers TEXT, PRIMARY KEY (id, host_name))",
			"CREATE TABLE IF NOT EXISTS tags (host_name TEXT REFERENCES hosts(name) ON DELETE CASCADE, tag TEXT, value TEXT, PRIMARY KEY (host_name, tag))",
			// "CREATE TABLE IF NOT EXISTS hosts (id TEXT PRIMARY KEY, name TEXT, data TEXT)",
		}

		wr, err := db.Write(statements)
		for _, v := range wr {
			if v.Err != nil {
				log.Errorf("sql syntax error: %s\n", v.Err.Error())
			}
		}
		if err != nil {
			return err
		}

		err = setVersion(db, 3)
		if err != nil {
			return err
		}
		version += 1
		log.Debugf("rqlite successfully updated tables to version 3.")
		fallthrough
	case 3:
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
	var statements []gorqlite.ParameterizedStatement

	for idx, host := range hosts {
		if host.Name == "" {
			return fmt.Errorf("host name required for host %d: %w", idx, ErrInvalidData)
		}

		// Keys are case-insensitive
		host.Name = strings.ToLower(host.Name)

		statements = append(statements, gorqlite.ParameterizedStatement{
			Query: "INSERT INTO hosts (name, provision, firmware, boot_image) VALUES(?, ?, ?, ?)",
			Arguments: []interface{}{
				host.Name,
				host.Provision,
				host.Firmware.String(),
				host.BootImage,
			},
		})

		for _, tag := range host.Tags {
			statements = append(statements, gorqlite.ParameterizedStatement{
				Query: "INSERT INTO tags (host_name, tag, value) VALUES(?, ?, ?)",
				Arguments: []interface{}{
					host.Name,
					tag,
					nil,
				},
			})
		}

		for i, iface := range host.Interfaces {
			statements = append(statements, gorqlite.ParameterizedStatement{
				Query: "INSERT INTO interfaces (id, host_name, mac, name, ip, fqdn, bmc, vlan, mtu) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)",
				Arguments: []interface{}{
					i,
					host.Name,
					iface.MAC.String(),
					iface.Name,
					iface.IP.String(),
					iface.FQDN,
					iface.BMC,
					iface.VLAN,
					iface.MTU,
				},
			})
		}
		for i, bonds := range host.Bonds {
			statements = append(statements, gorqlite.ParameterizedStatement{
				Query: "INSERT INTO interfaces (id, host_name, mac, name, ip, fqdn, bmc, vlan, mtu, peers) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				Arguments: []interface{}{
					i,
					host.Name,
					bonds.MAC.String(),
					bonds.Name,
					bonds.IP.String(),
					bonds.FQDN,
					bonds.BMC,
					bonds.VLAN,
					bonds.MTU,
					strings.Join(bonds.Peers, ","),
				},
			})
		}
	}

	wr, err := s.db.WriteParameterized(statements)
	for _, v := range wr {
		if v.Err != nil {
			fmt.Printf("SQL ERROR: %s\n", v.Err.Error())
		}
	}
	if err != nil {
		return err
	}

	log.Debugf("rqlite.StoreHosts: stored %d host(s)", len(hosts))

	return nil
}

// DeleteHosts deletes all hosts in the given nodeset.NodeSet from the data store.
func (s *RQLite) DeleteHosts(ns *nodeset.NodeSet) error {
	it := ns.Iterator()

	statements := []gorqlite.ParameterizedStatement{}

	for it.Next() {
		statements = append(statements, gorqlite.ParameterizedStatement{
			Query: "PRAGMA foreign_keys=on; DELETE FROM hosts WHERE name = ?",
			Arguments: []interface{}{
				it.Value(),
			},
		})
	}

	_, err := s.db.WriteParameterized(statements)
	if err != nil {
		return err
	}
	log.Debugf("rqlite.DeleteHosts: deleted %d host(s)", ns.Len())
	return nil
}

// LoadHostFromName returns the Host with the given name
func (s *RQLite) LoadHostFromName(name string) (*Host, error) {
	var host *Host

	hr, err := s.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
		Query: "SELECT * FROM hosts WHERE name = ?",
		Arguments: []interface{}{
			name,
		},
	})
	if err != nil {
		return nil, err
	}

	ir, err := s.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
		Query: "SELECT mac, name, ip, fqdn, bmc, vlan, mtu FROM interfaces WHERE host_name = ?",
		Arguments: []interface{}{
			name,
		},
	})
	if err != nil {
		return nil, err
	}

	br, err := s.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
		Query: "SELECT mac, name, ip, fqdn, bmc, vlan, mtu, peers FROM bonds WHERE host_name = ?",
		Arguments: []interface{}{
			name,
		},
	})
	if err != nil {
		return nil, err
	}

	tr, err := s.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
		Query: "SELECT tag FROM tags WHERE host_name = ?",
		Arguments: []interface{}{
			name,
		},
	})
	if err != nil {
		return nil, err
	}

	for hr.Next() {
		var fw string
		hr.Scan(&host.Name, &host.Provision, &fw, &host.BootImage)
		host.Firmware = firmware.NewFromString(fw)
	}
	for ir.Next() {
		var iface NetInterface
		var mac, ip string
		ir.Scan(&mac, &iface.Name, &ip, &iface.FQDN, &iface.BMC, &iface.VLAN, &iface.MTU)
		pm, err := net.ParseMAC(mac)
		if err != nil {
			log.Errorf("failed to parse MAC: %s", err)
		}
		pi, err := netip.ParsePrefix(ip)
		if err != nil {
			log.Errorf("failed to parse MAC: %s", err)
		}
		iface.MAC = pm
		iface.IP = pi
	}
	for br.Next() {
		var bond Bond
		var mac, ip, peers string
		br.Scan(&mac, &bond.Name, &ip, &bond.FQDN, &bond.BMC, &bond.VLAN, &bond.MTU, &peers)
		pm, err := net.ParseMAC(mac)
		if err != nil {
			log.Errorf("failed to parse MAC: %s", err)
		}
		pi, err := netip.ParsePrefix(ip)
		if err != nil {
			log.Errorf("failed to parse MAC: %s", err)
		}
		bond.MAC = pm
		bond.IP = pi
		bond.Peers = strings.Split(peers, ",")
	}
	for tr.Next() {
		var tag, value string
		tr.Scan(&tag, &value)
		// TODO: key-val pairs
		host.Tags = append(host.Tags, tag)
	}

	log.Debugf("%#v", host)

	log.Debugf("rqlite.LoadHostFromName: loaded host %s", name)
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

	hr, err := s.db.QueryOne("SELECT name, provision, firmware, boot_image FROM hosts")
	if err != nil {
		return nil, err
	}

	// TODO: Is is more efficient to send multiple queries per host vs query it all and sort it in GO?
	tr, err := s.db.QueryOne("SELECT host_name, tag, value FROM tags")
	if err != nil {
		return nil, err
	}

	var tags map[string][]string
	for tr.Next() {
		var hostName, tag, value string
		err := tr.Scan(&hostName, &tag, &value)
		if err != nil {
			tr.Next()
			log.Errorf("error scanning tags from DB: %s", err)
		}
		tags[hostName] = append(tags[hostName], tag)
	}

	for hr.Next() {
		var host Host
		var fw string

		err := hr.Scan(&host.Name, &host.Provision, &fw, &host.BootImage)
		if err != nil {
			hr.Next()
			log.Errorf("error scanning host from DB: %s", err)
		}
		host.Firmware = firmware.NewFromString(fw)

		host.Tags = tags[host.Name]

		hosts = append(hosts, &host)
	}

	log.Debugf("rqlite.Hosts: loaded %d host(s)", hr.NumRows())

	return hosts, nil
}

// FindHosts returns a list of all the hosts in the given NodeSet
func (s *RQLite) FindHosts(ns *nodeset.NodeSet) (HostList, error) {
	hosts := make(HostList, 0)

	it := ns.Iterator()

	for it.Next() {
		var host Host
		hr, err := s.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
			Query: "SELECT * FROM hosts WHERE name = ?",
			Arguments: []interface{}{
				it.Value(),
			},
		})
		if err != nil {
			return nil, err
		}

		ir, err := s.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
			Query: "SELECT mac, name, ip, fqdn, bmc, vlan, mtu FROM interfaces WHERE host_name = ?",
			Arguments: []interface{}{
				it.Value(),
			},
		})
		if err != nil {
			return nil, err
		}

		br, err := s.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
			Query: "SELECT mac, name, ip, fqdn, bmc, vlan, mtu, peers FROM bonds WHERE host_name = ?",
			Arguments: []interface{}{
				it.Value(),
			},
		})
		if err != nil {
			return nil, err
		}

		tr, err := s.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
			Query: "SELECT tag, value FROM tags WHERE host_name = ?",
			Arguments: []interface{}{
				it.Value(),
			},
		})
		if err != nil {
			return nil, err
		}

		for hr.Next() {
			var fw string
			hr.Scan(&host.Name, &host.Provision, &fw, &host.BootImage)
			host.Firmware = firmware.NewFromString(fw)
		}
		for ir.Next() {
			var iface NetInterface
			var mac, ip string
			ir.Scan(&mac, &iface.Name, &ip, &iface.FQDN, &iface.BMC, &iface.VLAN, &iface.MTU)
			pm, err := net.ParseMAC(mac)
			if err != nil {
				log.Errorf("failed to parse MAC: %s", err)
			}
			pi, err := netip.ParsePrefix(ip)
			if err != nil {
				log.Errorf("failed to parse MAC: %s", err)
			}
			iface.MAC = pm
			iface.IP = pi

			host.Interfaces = append(host.Interfaces, &iface)
		}
		for br.Next() {
			var bond Bond
			var mac, ip, peers string
			br.Scan(&mac, &bond.Name, &ip, &bond.FQDN, &bond.BMC, &bond.VLAN, &bond.MTU, &peers)
			pm, err := net.ParseMAC(mac)
			if err != nil {
				log.Errorf("failed to parse MAC: %s", err)
			}
			pi, err := netip.ParsePrefix(ip)
			if err != nil {
				log.Errorf("failed to parse MAC: %s", err)
			}
			bond.MAC = pm
			bond.IP = pi
			bond.Peers = strings.Split(peers, ",")

			host.Bonds = append(host.Bonds, &bond)
		}
		for tr.Next() {
			var tag, value string
			tr.Scan(&tag, &value)
			// TODO: key-val pairs
			host.Tags = append(host.Tags, tag)
		}

		log.Debugf("%#v", host)
		hosts = append(hosts, &host)

	}

	log.Debugf("rqlite.FindHosts: loaded host %s", ns.String())
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
