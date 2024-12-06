// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package buntstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

const (
	HostKeyPrefix      = "host"
	BootImageKeyPrefix = "image"
	UserKeyPrefix      = "user"
)

// BuntStore implements a Grendel Datastore using BuntDB
type BuntStore struct {
	db *buntdb.DB
}

type BuntUser struct {
	Username     string    `json:"username"`
	PasswordHash []byte    `json:"hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

// New returns a new BuntStore using the given database filename. For memory only you can provide `:memory:`
func New(filename string) (*BuntStore, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, err
	}

	err = db.CreateIndex("id", HostKeyPrefix+":*", buntdb.IndexJSON("id"))
	if err != nil && err != buntdb.ErrIndexExists {
		return nil, err
	}

	return &BuntStore{db: db}, nil
}

// Close closes the BuntStore database
func (s *BuntStore) Close() error {
	return s.db.Close()
}

// StoreUser stores the User in the data store
func (s *BuntStore) StoreUser(username, password string) (string, error) {
	d := true
	role := "disabled"

	err := s.db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(UserKeyPrefix + ":" + username)

		if err != nil && err == buntdb.ErrNotFound {
			d = false
		} else {
			return err
		}
		return nil
	})

	if err != nil {
		return role, err
	}
	if d {
		return role, fmt.Errorf("user %s already exists", username)
	}

	// Set role to admin if this is the first user
	count := 0
	err = s.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(UserKeyPrefix+":*", func(key, value string) bool {
			count++
			return true
		})
	})
	if err != nil {
		return role, err
	}
	if count == 0 {
		role = "admin"
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 8)
	return role, s.db.Update(func(tx *buntdb.Tx) error {
		user := BuntUser{
			PasswordHash: hashed,
			Role:         role,
			CreatedAt:    time.Now(),
			ModifiedAt:   time.Now(),
		}
		value, err := json.Marshal(user)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(UserKeyPrefix+":"+username, string(value), nil)
		if err != nil {
			return err
		}

		return nil
	})
}

// VerifyUser checks if the given username exists in the data store
func (s *BuntStore) VerifyUser(username, password string) (bool, string, error) {
	var dbVal BuntUser

	err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(UserKeyPrefix + ":" + username)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &dbVal)
	})

	if err != nil {
		return false, "", err
	}
	err = bcrypt.CompareHashAndPassword(dbVal.PasswordHash, []byte(password))
	if err != nil {
		return false, "", err
	}
	return true, dbVal.Role, nil
}

// GetUsers returns a list of all the usernames
func (s *BuntStore) GetUsers() ([]model.User, error) {
	var users []model.User
	var dbVal BuntUser

	err := s.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(UserKeyPrefix+":*", func(key, value string) bool {
			noPrefix := strings.Split(key, ":")[1]
			json.Unmarshal([]byte(value), &dbVal)

			users = append(users, model.User{Username: noPrefix, PasswordHash: string(dbVal.PasswordHash), Role: dbVal.Role, CreatedAt: dbVal.CreatedAt, ModifiedAt: dbVal.ModifiedAt})
			return true
		})
	})
	return users, err
}

// UpdateUserRole updates the role of the given users
func (s *BuntStore) UpdateUserRole(username, role string) error {
	var dbVal BuntUser

	err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(UserKeyPrefix + ":" + username)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &dbVal)
	})

	if err != nil {
		return err
	}
	dbVal.Role = role
	dbVal.ModifiedAt = time.Now()

	err = s.db.Update(func(tx *buntdb.Tx) error {
		val, err := json.Marshal(dbVal)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(UserKeyPrefix+":"+username, string(val), nil)
		return err
	})

	return err
}

// DeleteUser deletes the given user
func (s *BuntStore) DeleteUser(username string) error {
	err := s.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(UserKeyPrefix + ":" + username)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// StoreHost stores a host in the data store. If the host exists it is overwritten
func (s *BuntStore) StoreHost(host *model.Host) error {
	hostList := model.HostList{host}
	return s.StoreHosts(hostList)
}

// StoreHosts stores a list of host in the data store. If the host exists it is overwritten
func (s *BuntStore) StoreHosts(hosts model.HostList) error {
	for idx, host := range hosts {
		if host.Name == "" {
			return fmt.Errorf("host name required for host %d: %w", idx, store.ErrInvalidData)
		}

		// Keys are case-insensitive
		host.Name = strings.ToLower(host.Name)

		checkHost, err := s.LoadHostFromName(host.Name)
		if errors.Is(err, store.ErrNotFound) {
			uuid, err := ksuid.NewRandom()
			if err != nil {
				return err
			}

			host.UID = uuid
			continue
		}

		if err != nil {
			return fmt.Errorf("Failed to check host with name %s for duplicates:  %w", host.Name, err)
		}

		host.UID = checkHost.UID
	}

	err := s.db.Update(func(tx *buntdb.Tx) error {
		for _, host := range hosts {
			val, err := json.Marshal(host)
			if err != nil {
				return err
			}

			_, _, err = tx.Set(HostKeyPrefix+":"+host.Name, string(val), nil)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// DeleteHosts deletes all hosts in the given nodeset.NodeSet from the data store.
func (s *BuntStore) DeleteHosts(ns *nodeset.NodeSet) error {
	it := ns.Iterator()

	err := s.db.Update(func(tx *buntdb.Tx) error {
		for it.Next() {
			_, err := tx.Delete(HostKeyPrefix + ":" + it.Value())
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// LoadHostFromName returns the Host with the given name
func (s *BuntStore) LoadHostFromName(name string) (*model.Host, error) {
	var host *model.Host

	err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(HostKeyPrefix+":"+name, false)
		if err != nil {
			if err != buntdb.ErrNotFound {
				return err
			}

			return nil
		}

		host = &model.Host{}
		host.FromJSON(val)

		return nil
	})

	if err != nil {
		return nil, err
	}

	if host == nil {
		return nil, fmt.Errorf("host with name %s:  %w", name, store.ErrNotFound)
	}

	return host, nil
}

// LoadHostFromID returns the Host with the given ID
func (s *BuntStore) LoadHostFromID(id string) (*model.Host, error) {
	hostJSON := ""

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendEqual("id", fmt.Sprintf(`{"id":"%s"}`, id), func(key, value string) bool {
			hostJSON = value

			// XXX What to about dups? We only fetch first one.
			return false
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	if hostJSON == "" {
		return nil, fmt.Errorf("host with id %s:  %w", id, store.ErrNotFound)
	}

	host := &model.Host{}
	host.FromJSON(hostJSON)
	return host, nil
}

// ResolveIPv4 returns the list of IPv4 addresses with the given FQDN
func (s *BuntStore) ResolveIPv4(fqdn string) ([]net.IP, error) {
	fqdn = util.Normalize(fqdn)
	ips := make([]net.IP, 0)

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys(HostKeyPrefix+":*", func(key, value string) bool {
			for _, itype := range []string{"interfaces", "bonds"} {
				res := gjson.Get(value, itype)
				for _, i := range res.Array() {
					names := strings.Split(i.Get("fqdn").String(), ",")
					for _, name := range names {
						if util.Normalize(name) == fqdn {
							ip, _ := netip.ParsePrefix(i.Get("ip").String())
							if ip.IsValid() {
								ips = append(ips, net.IP(ip.Addr().AsSlice()))
							}

							// XXX stop after first match. consider changing this
							return false
						}
					}
				}
			}

			return true
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	return ips, nil
}

// ReverseResolve returns the list of FQDNs for the given IP
func (s *BuntStore) ReverseResolve(ip string) ([]string, error) {
	fqdn := make([]string, 0)

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys(HostKeyPrefix+":*", func(key, value string) bool {
			for _, itype := range []string{"interfaces", "bonds"} {
				res := gjson.Get(value, itype)
				for _, i := range res.Array() {
					ipWithMask := strings.Split(i.Get("ip").String(), "/")
					if len(ipWithMask) >= 1 && ipWithMask[0] == ip {
						names := strings.Split(i.Get("fqdn").String(), ",")
						for _, name := range names {
							fqdn = append(fqdn, name)
							break
						}

						// XXX stop after first match. consider changing this
						return false
					}
				}
			}

			return true
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	return fqdn, nil
}

// LoadHostFromMAC returns the Host that has a network interface with the give MAC address
func (s *BuntStore) LoadHostFromMAC(mac string) (*model.Host, error) {
	hostJSON := ""

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys(HostKeyPrefix+":*", func(key, value string) bool {
			res := gjson.Get(value, "interfaces")
			for _, i := range res.Array() {
				if i.Get("mac").String() == mac {
					hostJSON = value
					return false
				}
			}

			return true
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	if hostJSON == "" {
		return nil, fmt.Errorf("no host found with mac address %s:  %w", mac, store.ErrNotFound)
	}

	host := &model.Host{}
	host.FromJSON(hostJSON)
	return host, nil
}

// Hosts returns a list of all the hosts
func (s *BuntStore) Hosts() (model.HostList, error) {
	hosts := make(model.HostList, 0)

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys(HostKeyPrefix+":*", func(key, value string) bool {
			h := &model.Host{}
			h.FromJSON(value)
			hosts = append(hosts, h)
			return true
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	return hosts, nil
}

// FindHosts returns a list of all the hosts in the given NodeSet
func (s *BuntStore) FindHosts(ns *nodeset.NodeSet) (model.HostList, error) {
	hosts := make(model.HostList, 0)

	it := ns.Iterator()

	err := s.db.View(func(tx *buntdb.Tx) error {
		for it.Next() {
			val, err := tx.Get(HostKeyPrefix+":"+it.Value(), false)
			if err != nil {
				if err != buntdb.ErrNotFound {
					return err
				}
				continue
			}

			h := &model.Host{}
			h.FromJSON(val)
			hosts = append(hosts, h)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return hosts, nil
}

// FindTags returns a nodeset.NodeSet of all the hosts with the given tags
func (s *BuntStore) FindTags(tags []string) (*nodeset.NodeSet, error) {
	nodes := []string{}

	tagMap := make(map[string]struct{})
	for _, t := range tags {
		tagMap[t] = struct{}{}
	}

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys(HostKeyPrefix+":*", func(key, value string) bool {
			res := gjson.Get(value, "tags")
			for _, i := range res.Array() {
				if _, ok := tagMap[i.String()]; ok {
					nodes = append(nodes, gjson.Get(value, "name").String())
				}
			}

			return true
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no hosts found with tags %#v:  %w", tags, store.ErrNotFound)
	}

	return nodeset.NewNodeSet(strings.Join(nodes, ","))
}

// MatchTags returns a nodeset.NodeSet of all the hosts with the all given tags
func (s *BuntStore) MatchTags(tags []string) (*nodeset.NodeSet, error) {
	nodes := []string{}

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys(HostKeyPrefix+":*", func(key, value string) bool {
			res := gjson.Get(value, "tags")
			match := 0
			for _, v := range tags {
				for _, i := range res.Array() {
					if v == i.String() {
						match++
					}
				}
			}
			if match == len(tags) {
				nodes = append(nodes, gjson.Get(value, "name").String())
			}

			return true
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no hosts found with tags %#v:  %w", tags, store.ErrNotFound)
	}

	return nodeset.NewNodeSet(strings.Join(nodes, ","))
}

// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
func (s *BuntStore) ProvisionHosts(ns *nodeset.NodeSet, provision bool) error {
	it := ns.Iterator()
	count := 0

	err := s.db.Update(func(tx *buntdb.Tx) error {
		for it.Next() {
			key := HostKeyPrefix + ":" + it.Value()
			val, err := tx.Get(key, false)
			if err != nil {
				if err != buntdb.ErrNotFound {
					return err
				}
				continue
			}

			val, err = sjson.Set(val, "provision", provision)
			if err != nil {
				return err
			}

			_, _, err = tx.Set(key, val, nil)
			if err != nil {
				return err
			}

			count++
		}
		return nil
	})

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("no hosts found with nodeset %s:  %w", ns.String(), store.ErrNotFound)
	}

	return nil
}

// TagHosts adds tags to all hosts in the given NodeSet
func (s *BuntStore) TagHosts(ns *nodeset.NodeSet, tags []string) error {
	it := ns.Iterator()
	count := 0

	err := s.db.Update(func(tx *buntdb.Tx) error {
		for it.Next() {
			key := HostKeyPrefix + ":" + it.Value()
			val, err := tx.Get(key, false)
			if err != nil {
				if err != buntdb.ErrNotFound {
					return err
				}
				continue
			}

			uniqTags := make(map[string]struct{})
			res := gjson.Get(val, "tags")

			// Add existing  tags
			for _, i := range res.Array() {
				uniqTags[i.String()] = struct{}{}
			}

			// Add new tags
			for _, t := range tags {
				uniqTags[t] = struct{}{}
			}

			tagSlice := make([]string, 0, len(uniqTags))
			for v := range uniqTags {
				tagSlice = append(tagSlice, v)
			}

			val, err = sjson.Set(val, "tags", tagSlice)
			if err != nil {
				return err
			}

			_, _, err = tx.Set(key, val, nil)
			if err != nil {
				return err
			}

			count++
		}
		return nil
	})

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("no hosts found with nodeset %s:  %w", ns.String(), store.ErrNotFound)
	}

	return nil
}

// UntagHosts removes tags from all hosts in the given NodeSet
func (s *BuntStore) UntagHosts(ns *nodeset.NodeSet, tags []string) error {
	it := ns.Iterator()
	count := 0

	removeTags := make(map[string]struct{})
	for _, t := range tags {
		removeTags[t] = struct{}{}
	}

	err := s.db.Update(func(tx *buntdb.Tx) error {
		for it.Next() {
			key := HostKeyPrefix + ":" + it.Value()
			val, err := tx.Get(key, false)
			if err != nil {
				if err != buntdb.ErrNotFound {
					return err
				}
				continue
			}

			tagSlice := []string{}
			res := gjson.Get(val, "tags")
			for _, i := range res.Array() {
				if _, ok := removeTags[i.String()]; !ok {
					tagSlice = append(tagSlice, i.String())
				}
			}

			val, err = sjson.Set(val, "tags", tagSlice)
			if err != nil {
				return err
			}

			_, _, err = tx.Set(key, val, nil)
			if err != nil {
				return err
			}

			count++
		}
		return nil
	})

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("no hosts found with nodeset %s:  %w", ns.String(), store.ErrNotFound)
	}

	return nil
}

// SetBootImage sets all hosts to use the BootImage with the given name
func (s *BuntStore) SetBootImage(ns *nodeset.NodeSet, name string) error {
	it := ns.Iterator()

	err := s.db.Update(func(tx *buntdb.Tx) error {
		for it.Next() {
			key := HostKeyPrefix + ":" + it.Value()
			val, err := tx.Get(key, false)
			if err != nil {
				if err != buntdb.ErrNotFound {
					return err
				}
				continue
			}

			val, err = sjson.Set(val, "boot_image", name)
			if err != nil {
				return err
			}

			_, _, err = tx.Set(key, val, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// StoreBootImage stores a boot image in the data store. If the boot image exists it is overwritten
func (s *BuntStore) StoreBootImage(image *model.BootImage) error {
	imageList := model.BootImageList{image}
	return s.StoreBootImages(imageList)
}

// StoreBootImages stores a list of boot images in the data store. If the boot image exists it is overwritten
func (s *BuntStore) StoreBootImages(images model.BootImageList) error {
	for idx, image := range images {
		if image.Name == "" {
			return fmt.Errorf("name required for boot image %d: %w", idx, store.ErrInvalidData)
		}

		// Keys are case-insensitive
		image.Name = strings.ToLower(image.Name)

		// XXX need to check for dups?
		if image.UID.IsNil() {
			uuid, err := ksuid.NewRandom()
			if err != nil {
				return err
			}

			image.UID = uuid
		}
	}

	err := s.db.Update(func(tx *buntdb.Tx) error {
		for _, image := range images {
			val, err := json.Marshal(image)
			if err != nil {
				return err
			}

			_, _, err = tx.Set(BootImageKeyPrefix+":"+image.Name, string(val), nil)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// DeleteBootImages deletes boot images from the data store.
func (s *BuntStore) DeleteBootImages(names []string) error {
	err := s.db.Update(func(tx *buntdb.Tx) error {
		for _, name := range names {
			_, err := tx.Delete(BootImageKeyPrefix + ":" + name)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// LoadBootImage returns a BootImage with the given name
func (s *BuntStore) LoadBootImage(name string) (*model.BootImage, error) {
	var image *model.BootImage

	err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(BootImageKeyPrefix+":"+name, false)
		if err != nil {
			if err != buntdb.ErrNotFound {
				return err
			}

			return nil
		}

		var i model.BootImage
		err = json.Unmarshal([]byte(val), &i)
		if err != nil {
			return err
		}

		image = &i
		return nil
	})

	if err != nil {
		return nil, err
	}

	if image == nil {
		return nil, fmt.Errorf("boot image with name %s:  %w", name, store.ErrNotFound)
	}

	return image, nil
}

// BootImages returns a list of all boot images
func (s *BuntStore) BootImages() (model.BootImageList, error) {
	images := make(model.BootImageList, 0)

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys(BootImageKeyPrefix+":*", func(key, value string) bool {
			var i model.BootImage
			err := json.Unmarshal([]byte(value), &i)
			if err == nil {
				images = append(images, &i)
			} else {
				store.Log.WithFields(logrus.Fields{
					"err": err,
				}).Warn("Invalid boot image json stored in db")
			}
			return true
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	return images, nil
}

// RestoreFrom restores the database using the provided data dump
func (s *BuntStore) RestoreFrom(data model.DataDump) error {
	err := s.db.Update(func(tx *buntdb.Tx) error {
		for _, user := range data.Users {
			buser := &BuntUser{
				Username:     user.Username,
				PasswordHash: []byte(user.PasswordHash),
				Role:         user.Role,
				CreatedAt:    user.CreatedAt,
				ModifiedAt:   user.ModifiedAt,
			}
			val, err := json.Marshal(buser)
			if err != nil {
				return err
			}

			_, _, err = tx.Set(UserKeyPrefix+":"+user.Username, string(val), nil)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = s.StoreBootImages(data.Images)
	if err != nil {
		return err
	}

	return s.StoreHosts(data.Hosts)
}
