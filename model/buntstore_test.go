package model

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func tempfile() string {
	name, err := ioutil.TempFile("", "grendel-bunt-")
	if err != nil {
		panic(err)
	}
	return name.Name()
}

func TestBuntStoreHost(t *testing.T) {
	assert := assert.New(t)

	store, err := NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	badhost := &Host{}
	err = store.StoreHost(badhost)
	assert.Error(ErrInvalidData, err)

	host := HostFactory.MustCreate().(*Host)

	err = store.StoreHost(host)
	assert.NoError(err)

	_, err = store.LoadHostByID("notfound")
	assert.Error(ErrNotFound, err)

	_, err = store.LoadHostByName("notfound")
	assert.Error(ErrNotFound, err)

	testHost, err := store.LoadHostByID(host.ID.String())
	assert.Nil(err)
	assert.Equal(2, len(testHost.Interfaces))

	testHost2, err := store.LoadHostByName(host.Name)
	if assert.Nil(err) {
		assert.Equal(host.Name, testHost2.Name)
		assert.True(host.Interfaces[0].IP.Equal(testHost2.Interfaces[0].IP))
	}

	testIPs, err := store.LoadNetInterfaces(host.Interfaces[0].FQDN)
	if assert.Nil(err) {
		assert.Equal(1, len(testIPs))
		assert.Equal(host.Interfaces[0].IP.String(), testIPs[0].String())
	}
}

func BenchmarkBuntStoreWriteHost(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := NewBuntStore(file)
	defer store.Close()
	if err != nil {
		panic(err)
	}

	size := 5000
	for n := 0; n < b.N; n++ {
		for i := 0; i < size; i++ {
			host := HostFactory.MustCreate().(*Host)
			err := store.StoreHost(host)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkBuntStoreReadHost(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := NewBuntStore(file)
	defer store.Close()
	if err != nil {
		panic(err)
	}

	size := 5000
	rand.Seed(time.Now().UnixNano())
	hosts := make(HostList, size)
	for i := 0; i < size; i++ {
		host := HostFactory.MustCreate().(*Host)
		err = store.StoreHost(host)
		if err != nil {
			panic(err)
		}
		hosts[i] = host
	}

	b.SetParallelism(128)
	for n := 0; n < b.N; n++ {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				pick := hosts[rand.Intn(size)]
				_, err := store.LoadHostByID(pick.ID.String())
				if err != nil {
					panic(err)
				}

				_, err = store.LoadHostByName(pick.Name)
				if err != nil {
					panic(err)
				}

				ips, err := store.LoadNetInterfaces(pick.Interfaces[0].FQDN)
				if err != nil || len(ips) != 1 {
					panic(err)
				}
			}
		})

	}
}
