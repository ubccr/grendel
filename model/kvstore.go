package model

import (
	"github.com/segmentio/ksuid"
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

func (s *KVStore) GetHost(id string) (*Host, error) {
	uuid, err := ksuid.Parse(id)
	if err != nil {
		return nil, err
	}

	host := &Host{}

	err = s.store.Get(uuid.Bytes(), host)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func (s *KVStore) SaveHost(host *Host) error {
	if host.ID.IsNil() {
		uuid, err := ksuid.NewRandom()
		if err != nil {
			return err
		}
		host.ID = uuid
	}

	return s.store.Upsert(host.ID.Bytes(), host)
}

func (s *KVStore) HostList() (HostList, error) {
	var result HostList

	err := s.store.Find(&result, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *KVStore) Close() error {
	return s.store.Close()
}
