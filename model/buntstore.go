package model

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
)

type BuntStore struct {
	db *buntdb.DB
}

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

func (kv *BuntStore) Close() error {
	return kv.db.Close()
}

func (s *BuntStore) StoreHost(host *Host) error {
	if host.Name == "" {
		return fmt.Errorf("name required:  %w", ErrInvalidData)
	}

	val, err := json.Marshal(host)
	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("hosts:"+host.Name, string(val), nil)
		return err
	})

	return err
}

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
