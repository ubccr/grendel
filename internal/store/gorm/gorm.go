// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package gormstore

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type GormStore struct {
	db *gorm.DB
}

func New(filename string) (*GormStore, error) {
	conn := sqlite.Open(filename + "?mode=rwc&_journal_mode=WAL")

	db, err := gorm.Open(conn, &gorm.Config{
		CreateBatchSize: 1000,
		Logger:          logger.Discard,
		// SkipDefaultTransaction: true,
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	schema.RegisterSerializer("MACSerializer", &MACSerializer{})
	schema.RegisterSerializer("IPPrefixSerializer", &IPPrefixSerializer{})

	if err = db.AutoMigrate(&model.User{}, &model.Host{}, &model.NetInterface{}, &model.Bond{}, &model.BootImage{}); err != nil {
		store.Log.Error("db automigrate failed")
		return nil, err
	}

	return &GormStore{
		db: db,
	}, nil
}

// StoreUser stores the User in the data store
func (s *GormStore) StoreUser(username, password string) (string, error) {
	role := model.RoleDisabled

	var count int64
	s.db.Model(&model.User{}).Count(&count)

	if count == 0 {
		role = model.RoleAdmin
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", err
	}

	user := model.User{
		Username:     username,
		Role:         role.String(),
		PasswordHash: string(hashed),
		ModifiedAt:   time.Now(),
	}

	err = s.db.Create(&user).Error

	return role.String(), err
}

// VerifyUser checks if the given username exists in the data store
func (s *GormStore) VerifyUser(username, password string) (bool, string, error) {
	var user model.User

	err := s.db.Where(&model.User{Username: username}).First(&user).Error
	if err != nil {
		return false, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))

	return true, user.Role, err
}

// GetUsers returns a list of all the usernames
func (s *GormStore) GetUsers() ([]model.User, error) {
	var users []model.User

	err := s.db.Find(&users).Error

	return users, err
}

// UpdateUser updates the role of the given users
func (s *GormStore) UpdateUserRole(username, roleName string) error {
	var user model.User

	role, err := model.RoleFromString(roleName)
	if err != nil {
		return err
	}

	err = s.db.Where(&model.User{Username: username}).First(&user).Error
	if err != nil {
		return err
	}

	user.Role = role.String()

	return s.db.Save(&user).Error
}

// DeleteUser deletes the given user
func (s *GormStore) DeleteUser(username string) error {
	return s.db.Where(&model.User{Username: username}).Delete(&model.User{}).Error
}

// StoreHost stores a host in the data store. If the host exists it is overwritten
func (s *GormStore) StoreHost(host *model.Host) error {
	hostList := model.HostList{host}
	return s.StoreHosts(hostList)
}

// StoreHosts stores a list of host in the data store. If the host exists it is overwritten
func (s *GormStore) StoreHosts(hosts model.HostList) error {
	for idx, host := range hosts {
		if host.Name == "" {
			return fmt.Errorf("host name required for host %d: %w", idx, store.ErrInvalidData)
		}

		if host.UID.IsNil() {
			var err error
			host.UID, err = ksuid.NewRandom()
			if err != nil {
				return err
			}
		}
	}
	err := s.db.Clauses(
		clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true},
		clause.OnConflict{Columns: []clause.Column{{Name: "uid"}}, UpdateAll: true},
		clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, UpdateAll: true},
	).Create(&hosts).Error

	return err
}

// DeleteHosts deletes all hosts in the given nodeset.NodeSet from the data store.
func (s *GormStore) DeleteHosts(ns *nodeset.NodeSet) error {
	return s.db.Where("name IN ?", ns.Iterator().StringSlice()).Delete(&model.Host{}).Error
}

// LoadHostFromName returns the Host with the given name
func (s *GormStore) LoadHostFromName(name string) (*model.Host, error) {
	var host *model.Host

	err := s.db.Preload("Interfaces").Preload("Bonds").Where(&model.Host{Name: name}).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		return nil, store.ErrNotFound
	}

	return host, err
}

// LoadHostFromID returns the Host with the given ID
func (s *GormStore) LoadHostFromID(uid string) (*model.Host, error) {
	var host *model.Host

	err := s.db.Preload("Interfaces").Preload("Bonds").Where("uid = ?", uid).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		return nil, store.ErrNotFound
	}

	return host, err
}

// ResolveIPv4 returns the list of IPv4 addresses with the given FQDN
func (s *GormStore) ResolveIPv4(fqdn string) ([]net.IP, error) {
	var ifaces []*model.NetInterface
	var bonds []*model.Bond
	ips := make([]net.IP, 0)

	err := s.db.Where(&model.NetInterface{FQDN: fqdn}).Find(&ifaces).Error
	if err != nil {
		return ips, err
	}

	for _, iface := range ifaces {
		if iface.IP.IsValid() {
			ips = append(ips, net.IP(iface.IP.Addr().AsSlice()))
		}
	}

	err = s.db.Where(&model.Bond{NetInterface: model.NetInterface{FQDN: fqdn}}).Find(&bonds).Error
	if err != nil {
		return ips, err
	}

	for _, bond := range bonds {
		if bond.IP.IsValid() {
			ips = append(ips, net.IP(bond.IP.Addr().AsSlice()))
		}
	}

	return ips, err
}

// ReverseResolve returns the list of FQDNs for the given IP
func (s *GormStore) ReverseResolve(ip string) ([]string, error) {
	if len(ip) == 0 {
		return nil, errors.New("invalid ip")
	}

	var ifaces []*model.NetInterface
	var bonds []*model.Bond
	fqdns := make([]string, 0)

	err := s.db.Where("ip LIKE ?", ip+"%").Find(&ifaces).Error
	if err != nil {
		return fqdns, err
	}

	for _, iface := range ifaces {
		strSlice := strings.Split(iface.FQDN, ",")
		for _, fqdn := range strSlice {
			fqdns = append(fqdns, fqdn)
		}
	}

	err = s.db.Where("ip LIKE ?", ip+"%").Find(&bonds).Error
	if err != nil {
		return fqdns, err
	}

	for _, bond := range bonds {
		strSlice := strings.Split(bond.FQDN, ",")
		for _, fqdn := range strSlice {
			fqdns = append(fqdns, fqdn)
		}
	}

	return fqdns, err
}

// LoadHostFromMAC returns the Host that has a network interface with the give MAC address
func (s *GormStore) LoadHostFromMAC(macStr string) (*model.Host, error) {
	if len(macStr) == 0 {
		return nil, errors.New("invalid mac")
	}
	var iface *model.NetInterface
	var host *model.Host

	err := s.db.Where("mac = ?", macStr).First(&iface).Error
	if err == gorm.ErrRecordNotFound {
		return nil, store.ErrNotFound
	}

	err = s.db.Preload("Interfaces").Preload("Bonds").Where(&model.Host{ID: iface.ID}).First(&host).Error

	return host, err
}

// Hosts returns a list of all the hosts
func (s *GormStore) Hosts() (model.HostList, error) {
	var hosts model.HostList
	err := s.db.Preload("Interfaces").Preload("Bonds").Find(&hosts).Error

	return hosts, err
}

// FindHosts returns a list of all the hosts in the given NodeSet
func (s *GormStore) FindHosts(ns *nodeset.NodeSet) (model.HostList, error) {
	var hosts model.HostList

	err := s.db.Preload("Interfaces").Preload("Bonds").Where("name IN ?", ns.Iterator().StringSlice()).Find(&hosts).Error

	return hosts, err
}

// FindTags returns a nodeset.NodeSet of all the hosts with the given tags
func (s *GormStore) FindTags(tags []string) (*nodeset.NodeSet, error) {
	var nodes []string
	var hosts model.HostList

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
		return nil, store.ErrNotFound
	}

	nodeString := strings.Join(nodes, ",")

	return nodeset.NewNodeSet(nodeString)
}

func (s *GormStore) MatchTags(tags []string) (*nodeset.NodeSet, error) {
	return nil, nil
}

// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
func (s *GormStore) ProvisionHosts(ns *nodeset.NodeSet, provision bool) error {
	return s.db.Model(model.Host{}).Where("name IN ?", ns.Iterator().StringSlice()).Update("provision", true).Error
}

// TagHosts adds tags to all hosts in the given NodeSet
func (s *GormStore) TagHosts(ns *nodeset.NodeSet, tags []string) error {
	it := ns.Iterator()

	tx := s.db.Begin()
	defer tx.Rollback()

	for it.Next() {
		var h model.Host
		err := tx.Where(&model.Host{Name: it.Value()}).First(&h).Error
		if err != nil {
			return err
		}

		h.Tags = append(h.Tags, tags...)

		err = tx.Save(&h).Error
		if err != nil {
			return err
		}
	}

	return tx.Commit().Error
}

// UntagHosts removes tags from all hosts in the given NodeSet
func (s *GormStore) UntagHosts(ns *nodeset.NodeSet, tags []string) error {
	it := ns.Iterator()

	removeTags := make(map[string]struct{})
	for _, t := range tags {
		removeTags[t] = struct{}{}
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	for it.Next() {
		var h model.Host
		var tags []string
		err := tx.Where(&model.Host{Name: it.Value()}).First(&h).Error
		if err != nil && err != gorm.ErrRecordNotFound {
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
			return err
		}
	}

	return tx.Commit().Error
}

// SetBootImage sets all hosts to use the BootImage with the given name
func (s *GormStore) SetBootImage(ns *nodeset.NodeSet, name string) error {
	return s.db.Model(model.Host{}).Where("name IN ?", ns.Iterator().StringSlice()).Update("boot_image", name).Error
}

// StoreBootImage stores a boot image in the data store. If the boot image exists it is overwritten
func (s *GormStore) StoreBootImage(image *model.BootImage) error {
	imageList := model.BootImageList{image}
	return s.StoreBootImages(imageList)
}

// StoreBootImages stores a list of boot images in the data store. If the boot image exists it is overwritten
func (s *GormStore) StoreBootImages(images model.BootImageList) error {
	for idx, image := range images {
		if image.Name == "" {
			return fmt.Errorf("name required for kernel %d: %w", idx, store.ErrInvalidData)
		}

		// Keys are case-insensitive
		image.Name = strings.ToLower(image.Name)

		if image.UID.IsNil() {
			uuid, err := ksuid.NewRandom()
			if err != nil {
				return err
			}

			image.UID = uuid
		}
	}

	err := s.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&images).Error

	return err
}

// DeleteBootImages deletes boot images from the data store.
func (s *GormStore) DeleteBootImages(names []string) error {
	return s.db.Where("name IN ?", names).Delete(&model.BootImage{}).Error
}

// LoadBootImage returns a BootImage with the given name
func (s *GormStore) LoadBootImage(name string) (*model.BootImage, error) {
	var image *model.BootImage

	err := s.db.Where(&model.BootImage{Name: name}).First(&image).Error
	if err == gorm.ErrRecordNotFound {
		err = store.ErrNotFound
	}

	return image, err
}

// BootImages returns a list of all boot images
func (s *GormStore) BootImages() (model.BootImageList, error) {
	var images model.BootImageList

	err := s.db.Find(&images).Error

	return images, err
}

func (s *GormStore) RestoreFrom(data model.DataDump) error {
	err := s.db.Save(&data.Users).Error
	if err != nil {
		return err
	}

	err = s.StoreBootImages(data.Images)
	if err != nil {
		return err
	}

	return s.StoreHosts(data.Hosts)
}

// Close closes the GORM database
func (s *GormStore) Close() error {
	raw, err := s.db.DB()
	if err != nil {
		return err
	}

	return raw.Close()
}
