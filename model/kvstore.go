package model

import (
	"net"

	"github.com/timshannon/badgerhold"
)

type KVStore struct {
	store *badgerhold.Store
}

func NewKVStore(filename string) (*KVStore, error) {
	options := badgerhold.DefaultOptions
	options.Dir = filename
	options.ValueDir = filename
	options.Logger = log
	store, err := badgerhold.Open(options)
	if err != nil {
		return nil, err
	}

	return &KVStore{store: store}, nil
}

func (s *KVStore) GetBootImage(mac string) (*BootImage, error) {
	return nil, nil
}

func (s *KVStore) GetHost(mac string) (*Host, error) {
	hwaddr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	host := &Host{}

	err = s.store.Get(hwaddr, host)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func (s *KVStore) SaveHost(host *Host) error {
	return s.store.Upsert(host.MAC, host)
}

func (s *KVStore) Close() error {
	return s.store.Close()
}
