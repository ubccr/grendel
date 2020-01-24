package model

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/ubccr/grendel/nodeset"
)

// BuntStore implements a Grendel Datastore using BuntDB
type BuntStore struct {
	db *buntdb.DB
}

// NewBuntStore returns a new BuntStore using the given database filename. For memory only you can provide `:memory:`
func NewBuntStore(filename string) (*BuntStore, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, err
	}

	err = db.CreateIndex("id", "hosts:*", buntdb.IndexJSON("id"))
	if err != nil && err != buntdb.ErrIndexExists {
		return nil, err
	}

	return &BuntStore{db: db}, nil
}

// Close closes the BuntStore database
func (s *BuntStore) Close() error {
	return s.db.Close()
}

// StoreHost stores a host in the data store. If the host exists it is overwritten
func (s *BuntStore) StoreHost(host *Host) error {
	if host.Name == "" {
		return fmt.Errorf("name required:  %w", ErrInvalidData)
	}

	// Keys are case-insensitive
	host.Name = strings.ToLower(host.Name)

	if host.ID.IsNil() {
		uuid, err := ksuid.NewRandom()
		if err != nil {
			return err
		}

		host.ID = uuid
	}

	val, err := json.Marshal(host)
	if err != nil {
		return err
	}

	hostJSON := string(val)

	err = s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("hosts:"+host.Name, hostJSON, nil)
		return err
	})

	if err != nil {
		return err
	}

	return err
}

// LoadHostByName returns the Host with the given name
func (s *BuntStore) LoadHostByName(name string) (*Host, error) {
	var host *Host

	err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get("hosts:"+name, false)
		if err != nil {
			if err != buntdb.ErrNotFound {
				return err
			}

			return nil
		}

		var h Host
		err = json.Unmarshal([]byte(val), &h)
		if err != nil {
			return err
		}

		host = &h
		return nil
	})

	if err != nil {
		return nil, err
	}

	if host == nil {
		return nil, fmt.Errorf("host with name %s:  %w", name, ErrNotFound)
	}

	return host, nil
}

// LoadHostByID returns the Host with the given ID
func (s *BuntStore) LoadHostByID(id string) (*Host, error) {
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
		return nil, fmt.Errorf("host with id %s:  %w", id, ErrNotFound)
	}

	var host Host
	err = json.Unmarshal([]byte(hostJSON), &host)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

// LoadNetInterfaces returns the list of IPs with the given FQDN
func (s *BuntStore) LoadNetInterfaces(fqdn string) ([]net.IP, error) {
	ips := make([]net.IP, 0)

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys("hosts:*", func(key, value string) bool {
			res := gjson.Get(value, "interfaces")
			for _, i := range res.Array() {
				if i.Get("fqdn").String() == fqdn {
					ip := net.ParseIP(i.Get("ip").String())
					if ip != nil && ip.To4() != nil {
						ips = append(ips, ip)
					}
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

	if len(ips) == 0 {
		return nil, fmt.Errorf("interfaces with fqdn %s:  %w", fqdn, ErrNotFound)
	}

	return ips, nil
}

// Hosts returns a list of all the hosts
func (s *BuntStore) Hosts() (HostList, error) {
	hosts := make(HostList, 0)

	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys("hosts:*", func(key, value string) bool {
			var h Host
			err := json.Unmarshal([]byte(value), &h)
			if err == nil {
				hosts = append(hosts, &h)
			} else {
				log.WithFields(logrus.Fields{
					"err": err,
				}).Warn("Invalid host json stored in db")
			}
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
func (s *BuntStore) FindHosts(ns *nodeset.NodeSet) (HostList, error) {
	hosts := make(HostList, 0)

	it := ns.Iterator()

	err := s.db.View(func(tx *buntdb.Tx) error {
		for it.Next() {
			val, err := tx.Get("hosts:"+it.Value(), false)
			if err != nil {
				if err != buntdb.ErrNotFound {
					return err
				}
				continue
			}

			var h Host
			err = json.Unmarshal([]byte(val), &h)
			if err != nil {
				return err
			}

			hosts = append(hosts, &h)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return hosts, nil
}

// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
func (s *BuntStore) ProvisionHosts(ns *nodeset.NodeSet, provision bool) error {
	it := ns.Iterator()

	err := s.db.Update(func(tx *buntdb.Tx) error {
		for it.Next() {
			key := "hosts:" + it.Value()
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
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
