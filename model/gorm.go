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
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	// _ "github.com/rqlite/gorqlite/stdlib"
	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/nodeset"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// GORM implements a Grendel Datastore using SQLite or RQLite
type GORM struct {
	sqldb *sql.DB
	db    *gorm.DB
}

// NewGORMStore returns a new GORM using the given database address.
func NewGORMStore(dbType, path, addr string) (*GORM, error) {
	var conn gorm.Dialector

	switch dbType {
	case "sqlite":
		conn = sqlite.Open(path + "?mode=rwc&_journal_mode=WAL") // &_sync=NORMAL ?? seems to reduce performance by a few percent? Margin of error?
	default:
		return nil, errors.New("invalid database type. supported gorm options: sqlite")
	}

	db, err := gorm.Open(conn, &gorm.Config{
		CreateBatchSize: 1000,
		Logger:          logger.Discard,
		// SkipDefaultTransaction: true,
		PrepareStmt: true,
	})
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
		log.Error("db automigrate failed")
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
func (s *GORM) StoreUser(username, password string) (string, error) {
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
	return u.Role, err
}

// VerifyUser checks if the given username exists in the data store
func (s *GORM) VerifyUser(username, password string) (bool, string, error) {
	var u User
	if err := s.db.Where(&User{Username: username}).First(&u).Error; err != nil {
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
	err := s.db.Where(&User{Username: username}).Delete(&u).Error

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
	for idx, host := range hosts {
		if host.Name == "" {
			return fmt.Errorf("host name required for host %d: %w", idx, ErrInvalidData)
		}

		// Keys are case-insensitive
		host.Name = strings.ToLower(host.Name)

		// Always generate a new ksuid to avoid duplicates
		uuid, err := ksuid.NewRandom()
		if err != nil {
			return err
		}

		host.ID = uuid
	}
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
	var hostNames []string
	for it.Next() {
		hostNames = append(hostNames, it.Value())
	}

	err := s.db.Where("name IN ?", hostNames).Delete(&Host{}).Error

	log.Debugf("GORM.DeleteHosts: deleted %d host(s)", ns.Len())
	return err
}

// LoadHostFromName returns the Host with the given name
func (s *GORM) LoadHostFromName(name string) (*Host, error) {
	var host *Host
	err := s.db.Preload("Interfaces").Preload("Bonds").Where(&Host{Name: name}).First(&host).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("host with name %s:  %w", name, ErrNotFound)
	}

	log.Debugf("GORM.LoadHostFromName: loaded host %s", name)
	return host, err
}

// LoadHostFromID returns the Host with the given ID
func (s *GORM) LoadHostFromID(id string) (*Host, error) {
	var host *Host
	err := s.db.Preload("Interfaces").Preload("Bonds").Where("id = ?", id).First(&host).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("host with ID %s:  %w", id, ErrNotFound)
	}

	log.Debugf("GORM.LoadHostFromID: loaded host %s", host.Name)
	return host, err
}

// ResolveIPv4 returns the list of IPv4 addresses with the given FQDN
func (s *GORM) ResolveIPv4(fqdn string) ([]net.IP, error) {
	var ifaces []*NetInterface
	var bonds []*Bond
	ips := make([]net.IP, 0)

	err := s.db.Where(&NetInterface{FQDN: fqdn}).Find(&ifaces).Error
	if err != nil {
		return ips, err
	}

	for _, iface := range ifaces {
		if iface.IP.IsValid() {
			ips = append(ips, net.IP(iface.IP.Addr().AsSlice()))
		}
	}

	err = s.db.Where(&Bond{NetInterface: NetInterface{FQDN: fqdn}}).Find(&bonds).Error
	if err != nil {
		return ips, err
	}

	for _, bond := range bonds {
		if bond.IP.IsValid() {
			ips = append(ips, net.IP(bond.IP.Addr().AsSlice()))
		}
	}

	log.Debugf("GORM.ResolveIPv4: resolved %d ip(s)", len(ips))
	return ips, err
}

// ReverseResolve returns the list of FQDNs for the given IP
func (s *GORM) ReverseResolve(ip string) ([]string, error) {
	var ifaces []*NetInterface
	var bonds []*Bond
	fqdns := make([]string, 0)

	err := s.db.Where("ip LIKE ?", "%"+ip+"%").Find(&ifaces).Error
	if err != nil {
		return fqdns, err
	}

	for _, iface := range ifaces {
		fqdns = append(fqdns, iface.FQDN)
	}

	err = s.db.Where("ip LIKE ?", "%"+ip+"%").Find(&bonds).Error
	if err != nil {
		return fqdns, err
	}

	for _, bond := range bonds {
		fqdns = append(fqdns, bond.FQDN)
	}

	log.Debugf("GORM.ReverseResolve: resolved %d fqdn(s)", len(fqdns))
	return fqdns, err
}

// LoadHostFromMAC returns the Host that has a network interface with the give MAC address
func (s *GORM) LoadHostFromMAC(macStr string) (*Host, error) {
	var iface *NetInterface
	var host *Host

	err := s.db.Where("mac = ?", macStr).First(&iface).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("host with MAC %s:  %w", macStr, ErrNotFound)
	}

	err = s.db.Preload("Interfaces").Preload("Bonds").Where(&Host{GormID: iface.HostID}).First(&host).Error

	log.Debugf("GORM.LoadHostFromMAC: retrieved host %s from mac %s", host.Name, macStr)
	return host, err
}

// Hosts returns a list of all the hosts
func (s *GORM) Hosts() (HostList, error) {
	var hosts HostList
	err := s.db.Preload("Interfaces").Preload("Bonds").Find(&hosts).Error

	log.Debugf("GORM.Hosts: retrieved %d host(s)", len(hosts))
	return hosts, err
}

// FindHosts returns a list of all the hosts in the given NodeSet
func (s *GORM) FindHosts(ns *nodeset.NodeSet) (HostList, error) {
	var nodes []string
	var hosts HostList
	it := ns.Iterator()

	for it.Next() {
		nodes = append(nodes, it.Value())
	}

	err := s.db.Preload("Interfaces").Preload("Bonds").Where("name IN ?", nodes).Find(&hosts).Error
	if err != nil {
		return hosts, err
	}

	log.Debugf("GORM.FindHosts: found host %s", ns.String())
	return hosts, nil
}

// FindTags returns a nodeset.NodeSet of all the hosts with the given tags
func (s *GORM) FindTags(tags []string) (*nodeset.NodeSet, error) {
	var nodes []string
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

	nodeString := strings.Join(nodes, ",")

	log.Debugf("GORM.FindTags: loaded host(s) %s tagged with %s", nodeString, tags)
	return nodeset.NewNodeSet(nodeString)
}

// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
func (s *GORM) ProvisionHosts(ns *nodeset.NodeSet, provision bool) error {
	hostNames := []string{}
	it := ns.Iterator()

	for it.Next() {
		hostNames = append(hostNames, it.Value())
	}

	err := s.db.Model(Host{}).Where("name IN ?", hostNames).Update("provision", true).Error

	log.Debugf("GORM.ProvisionHosts: set host(s) %s to provision=%t", ns.String(), provision)
	return err
}

// TagHosts adds tags to all hosts in the given NodeSet
func (s *GORM) TagHosts(ns *nodeset.NodeSet, tags []string) error {
	it := ns.Iterator()

	tx := s.db.Begin()

	for it.Next() {
		var h Host
		err := tx.Where(&Host{Name: it.Value()}).First(&h).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		h.Tags = append(h.Tags, tags...)

		err = tx.Save(&h).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	log.Debugf("GORM.TagHosts: added tag(s) %s to host(s) %s", tags, ns.String())
	return tx.Commit().Error
}

// UntagHosts removes tags from all hosts in the given NodeSet
func (s *GORM) UntagHosts(ns *nodeset.NodeSet, tags []string) error {
	it := ns.Iterator()

	removeTags := make(map[string]struct{})
	for _, t := range tags {
		removeTags[t] = struct{}{}
	}

	tx := s.db.Begin()

	for it.Next() {
		var h Host
		var tags []string
		err := tx.Where(&Host{Name: it.Value()}).First(&h).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return err
		}
		for _, t := range h.Tags {
			if _, ok := removeTags[t]; !ok {
				tags = append(tags, t)
			}
		}
		h.Tags = tags
		err = tx.Save(&h).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	log.Debugf("GORM.UntagHosts: removed tag(s) %s from host(s) %s", tags, ns.String())
	return tx.Commit().Error
}

// SetBootImage sets all hosts to use the BootImage with the given name
func (s *GORM) SetBootImage(ns *nodeset.NodeSet, name string) error {
	hostNames := []string{}
	it := ns.Iterator()

	for it.Next() {
		hostNames = append(hostNames, it.Value())
	}

	err := s.db.Model(Host{}).Where("name IN ?", hostNames).Update("boot_image", name).Error

	log.Debugf("GORM.SetBootImage: set host(s) %s to use boot image %s", ns.String(), name)
	return err
}

// StoreBootImage stores a boot image in the data store. If the boot image exists it is overwritten
func (s *GORM) StoreBootImage(image *BootImage) error {
	imageList := BootImageList{image}
	return s.StoreBootImages(imageList)
}

// StoreBootImages stores a list of boot images in the data store. If the boot image exists it is overwritten
func (s *GORM) StoreBootImages(images BootImageList) error {
	// TODO: this is not specific to the DB and should be done outside of it?
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
	err := s.db.Where("name IN ?", names).Delete(&BootImage{}).Error

	log.Debugf("GORM.DeleteBootImages: deleted %d image(s)", len(names))
	return err
}

// LoadBootImage returns a BootImage with the given name
func (s *GORM) LoadBootImage(name string) (*BootImage, error) {
	var image *BootImage

	err := s.db.Where(&BootImage{Name: name}).First(&image).Error
	if err == gorm.ErrRecordNotFound {
		err = ErrNotFound
	}

	log.Debugf("GORM.LoadBootImages: loaded image %s", image.Name)
	return image, err
}

// BootImages returns a list of all boot images
func (s *GORM) BootImages() (BootImageList, error) {
	var images BootImageList

	err := s.db.Find(&images).Error

	log.Debugf("GORM.BootImages: loaded %d image(s)", len(images))
	return images, err
}
