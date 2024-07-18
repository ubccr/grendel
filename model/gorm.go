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
	"database/sql"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/jafurlan/rqlite"
	_ "github.com/rqlite/gorqlite/stdlib"
	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// GORM implements a Grendel Datastore using BuntDB
type GORM struct {
	sqldb *sql.DB
	db    *gorm.DB
}

// NewGORMStore returns a new GORM using the given database address.
func NewGORMStore(dbType, path, addr string) (*GORM, error) {
	var conn gorm.Dialector

	switch dbType {
	case "sqlite":
		conn = sqlite.Open(path)
	case "rqlite":
		conn = rqlite.Open(addr)
	}

	db, err := gorm.Open(conn, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqldb, err := db.DB()
	if err != nil {
		return nil, err
	}

	schema.RegisterSerializer("MACSerializer", &MACSerializer{})
	schema.RegisterSerializer("IPPrefixSerializer", &IPPrefixSerializer{})

	if err = db.AutoMigrate(&User{}, &Host{}, &NetInterface{}, &BootImage{}, &Bond{}); err != nil {
		return nil, err
	}
	return &GORM{
		sqldb: sqldb,
		db:    db,
	}, nil
}

// Close closes the GORM database
func (s *GORM) Close() error {
	return s.sqldb.Close()
}

// StoreUser stores the User in the data store
func (s *GORM) StoreUser(username, password string) error {
	role := "disabled"

	// Set role to admin if this is the first user
	var c int64
	s.db.Model(&User{}).Count(&c)

	if c == 0 {
		role = "admin"
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 8)
	u := User{
		Username:   username,
		Role:       role,
		Hash:       string(hashed),
		ModifiedAt: time.Now(),
	}
	err := s.db.Create(&u).Error

	log.Debugf("GORM.StoreUser: inserting '%s' user", username)
	return err
}

// VerifyUser checks if the given username exists in the data store
func (s *GORM) VerifyUser(username, password string) (bool, string, error) {
	var u User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		return false, "", err
	}

	log.Debugf("GORM.VerifyUser: queried user: %s", username)
	err := bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password))

	return true, u.Role, err
}

// GetUsers returns a list of all the usernames
func (s *GORM) GetUsers() ([]User, error) {
	var users []User

	err := s.db.Find(&users).Error

	log.Debugf("GORM.GetUsers: queried %d user(s)", len(users))
	return users, err
}

// UpdateUser updates the role of the given users
func (s *GORM) UpdateUser(username, role string) error {
	var u User

	if err := s.db.Where(&User{Username: username}).First(&u).Error; err != nil {
		return err
	}

	u.Role = role
	err := s.db.Save(&u).Error

	log.Debugf("GORM.UpdateUser: updated user: %s role", username)
	return err
}

// DeleteUser deletes the given user
func (s *GORM) DeleteUser(username string) error {
	var u User
	err := s.db.Where(&User{Username: username}, username).Find(&u).Delete(&u).Error

	log.Debugf("GORM.DeleteUser: deleting %s user", username)
	return err
}

// StoreHost stores a host in the data store. If the host exists it is overwritten
func (s *GORM) StoreHost(host *Host) error {
	hostList := HostList{host}
	return s.StoreHosts(hostList)
}

// StoreHosts stores a list of host in the data store. If the host exists it is overwritten
func (s *GORM) StoreHosts(hosts HostList) error {
	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).Create(&hosts).Error

	log.Debugf("GORM.StoreHosts: stored %d host(s)", len(hosts))
	return err
}

// DeleteHosts deletes all hosts in the given nodeset.NodeSet from the data store.
func (s *GORM) DeleteHosts(ns *nodeset.NodeSet) error {
	it := ns.Iterator()
	err := s.db.Delete(&Host{Name: it.Value()}).Error

	log.Debugf("GORM.DeleteHosts: deleted %d host(s)", ns.Len())
	return err
}

// LoadHostFromName returns the Host with the given name
func (s *GORM) LoadHostFromName(name string) (*Host, error) {
	var host *Host
	err := s.db.Where(&Host{Name: name}).Find(&host).Error

	log.Debugf("GORM.LoadHostFromName: loaded host %s", name)
	return host, err
}

// LoadHostFromID returns the Host with the given ID
func (s *GORM) LoadHostFromID(id string) (*Host, error) {
	kid, err := ksuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var host *Host
	err = s.db.Where(&Host{ID: kid}).First(&host).Error

	log.Debugf("GORM.LoadHostFromName: loaded host %s", host.Name)

	return host, err
}

// ResolveIPv4 returns the list of IPv4 addresses with the given FQDN
func (s *GORM) ResolveIPv4(fqdn string) ([]net.IP, error) {
	var hosts HostList
	fqdn = util.Normalize(fqdn) // TODO: ???
	ips := make([]net.IP, 0)

	err := s.db.Preload("Interfaces").Where(&NetInterface{FQDN: fqdn}).Find(&hosts).Error

	for _, h := range hosts {
		for _, i := range h.Interfaces {
			ip, _ := netip.ParsePrefix(i.IP.String())
			if ip.IsValid() {
				ips = append(ips, net.IP(ip.Addr().AsSlice()))
			}
		}
	}

	return ips, err
}

// ReverseResolve returns the list of FQDNs for the given IP
func (s *GORM) ReverseResolve(ip string) ([]string, error) {
	var hosts HostList
	fqdns := make([]string, 0)
	prefix, err := netip.ParsePrefix(ip)
	if err != nil {
		return []string{}, nil
	}

	err = s.db.Preload("Interfaces").Where(&NetInterface{IP: prefix}).Find(&hosts).Error

	for _, h := range hosts {
		for _, i := range h.Interfaces {
			fqdns = append(fqdns, strings.Split(i.FQDN, ",")...)
		}
	}

	return fqdns, err
}

// LoadHostFromMAC returns the Host that has a network interface with the give MAC address
func (s *GORM) LoadHostFromMAC(macStr string) (*Host, error) {
	var host *Host

	mac, err := net.ParseMAC(macStr)
	if err != nil {
		return nil, err
	}
	err = s.db.Preload("Interfaces").Where(&NetInterface{MAC: mac}).First(&host).Error

	return host, err
}

// Hosts returns a list of all the hosts
func (s *GORM) Hosts() (HostList, error) {
	var hosts HostList
	err := s.db.Preload("Interfaces").Preload("Bonds").Find(&hosts).Error

	log.Debugf("GORM.Hosts: loaded %d host(s)", len(hosts))

	return hosts, err
}

// FindHosts returns a list of all the hosts in the given NodeSet
func (s *GORM) FindHosts(ns *nodeset.NodeSet) (HostList, error) {
	hosts := make(HostList, 0)
	it := ns.Iterator()

	for it.Next() {
		var h Host
		err := s.db.Preload("Interfaces").Preload("Bonds").Where(&Host{Name: it.Value()}).Find(&h).Error
		if err != nil {
			log.Error(err)
		}
		hosts = append(hosts, &h)
	}
	log.Debugf("GORM.FindHosts: loaded host %s", ns.String())
	return hosts, nil
}

// FindTags returns a nodeset.NodeSet of all the hosts with the given tags
func (s *GORM) FindTags(tags []string) (*nodeset.NodeSet, error) {
	nodes := []string{}
	var hosts HostList

	tagMap := make(map[string]struct{})
	for _, t := range tags {
		tagMap[t] = struct{}{}
	}

	err := s.db.Find(&hosts).Error
	if err != nil {
		return nil, err
	}

	// TODO: converting to key/value pairs might fix not being able to filter with a Where statement
	for _, h := range hosts {
		for _, t := range h.Tags {
			if _, ok := tagMap[t]; ok {
				nodes = append(nodes, h.Name)
			}
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no hosts found with tags %#v:  %w", tags, ErrNotFound)
	}

	return nodeset.NewNodeSet(strings.Join(nodes, ","))
}

// MatchTags returns a nodeset.NodeSet of all the hosts with the all given tags
func (s *GORM) MatchTags(tags []string) (*nodeset.NodeSet, error) {
	var hosts HostList

	err := s.db.Where(&Host{Tags: tags}).Find(&hosts).Error
	if err != nil {
		return nil, err
	}

	return hosts.ToNodeSet()
}

// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
func (s *GORM) ProvisionHosts(ns *nodeset.NodeSet, provision bool) error {
	it := ns.Iterator()

	for it.Next() {
		err := s.db.Where(&Host{Name: it.Value()}).Save(&Host{Provision: provision}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// TagHosts adds tags to all hosts in the given NodeSet
func (s *GORM) TagHosts(ns *nodeset.NodeSet, tags []string) error {
	it := ns.Iterator()

	for it.Next() {
		var h Host
		err := s.db.Where(&Host{Name: it.Value()}).First(&h).Save(&Host{Tags: append(h.Tags, tags...)}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// UntagHosts removes tags from all hosts in the given NodeSet
func (s *GORM) UntagHosts(ns *nodeset.NodeSet, tags []string) error {
	it := ns.Iterator()

	removeTags := make(map[string]struct{})
	for _, t := range tags {
		removeTags[t] = struct{}{}
	}

	for it.Next() {
		var h Host
		var tags []string
		err := s.db.Where(&Host{Name: it.Value()}).First(&h).Error
		if err != nil {
			return err
		}
		for _, t := range h.Tags {
			if _, ok := removeTags[t]; !ok {
				tags = append(tags, t)
			}
		}
		err = s.db.Save(&h).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// SetBootImage sets all hosts to use the BootImage with the given name
func (s *GORM) SetBootImage(ns *nodeset.NodeSet, name string) error {
	it := ns.Iterator()

	for it.Next() {
		var h Host
		err := s.db.Where(&Host{Name: it.Value()}).First(&h).Save(&Host{BootImage: name}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// StoreBootImage stores a boot image in the data store. If the boot image exists it is overwritten
func (s *GORM) StoreBootImage(image *BootImage) error {
	imageList := BootImageList{image}
	return s.StoreBootImages(imageList)
}

// StoreBootImages stores a list of boot images in the data store. If the boot image exists it is overwritten
func (s *GORM) StoreBootImages(images BootImageList) error {
	// TODO: this is not specific to the DB and should be done outside of it
	for idx, image := range images {
		if image.Name == "" {
			return fmt.Errorf("name required for boot image %d: %w", idx, ErrInvalidData)
		}

		// Keys are case-insensitive
		image.Name = strings.ToLower(image.Name)

		// XXX need to check for dups?
		if image.ID.IsNil() {
			uuid, err := ksuid.NewRandom()
			if err != nil {
				return err
			}

			image.ID = uuid
		}
	}

	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).Create(&images).Error

	log.Debugf("GORM.StoreBootImages: stored %d image(s)", len(images))
	return err
}

// DeleteBootImages deletes boot images from the data store.
func (s *GORM) DeleteBootImages(names []string) error {
	// TODO: soft delete??
	err := s.db.Delete("name IN ?", names).Error
	return err
}

// LoadBootImage returns a BootImage with the given name
func (s *GORM) LoadBootImage(name string) (*BootImage, error) {
	var image *BootImage

	err := s.db.Where(&BootImage{Name: name}).First(&image).Error

	return image, err
}

// BootImages returns a list of all boot images
func (s *GORM) BootImages() (BootImageList, error) {
	var images BootImageList

	err := s.db.Find(&images).Error

	return images, err
}
