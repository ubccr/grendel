package model

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/nodeset"
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

	host := HostFactory.MustCreate().(*Host)

	err = store.StoreHost(host)
	assert.NoError(err)

	testHost, err := store.LoadHostByID(host.ID.String())
	if assert.NoError(err) {
		assert.Equal(2, len(testHost.Interfaces))
	}

	testHost2, err := store.LoadHostByName(host.Name)
	if assert.NoError(err) {
		assert.Equal(host.Name, testHost2.Name)
		assert.True(host.Interfaces[0].IP.Equal(testHost2.Interfaces[0].IP))
	}

	testHost3, err := store.LoadHostByMAC(host.Interfaces[0].MAC.String())
	if assert.NoError(err) {
		assert.Equal(host.Name, testHost3.Name)
		assert.Equal(host.Interfaces[0].MAC.String(), testHost3.Interfaces[0].MAC.String())
	}

	testIPs, err := store.LoadNetInterfaces(host.Interfaces[0].FQDN)
	if assert.NoError(err) {
		assert.Equal(1, len(testIPs))
		assert.Equal(host.Interfaces[0].IP.String(), testIPs[0].String())
	}

	badhost := &Host{}
	err = store.StoreHost(badhost)
	if assert.Error(err) {
		assert.True(errors.Is(err, ErrInvalidData))
	}

	_, err = store.LoadHostByID("notfound")
	if assert.Error(err) {
		assert.True(errors.Is(err, ErrNotFound))
	}

	_, err = store.LoadHostByName("notfound")
	if assert.Error(err) {
		assert.True(errors.Is(err, ErrNotFound))
	}

	_, err = store.LoadHostByMAC("notfound")
	if assert.Error(err) {
		assert.True(errors.Is(err, ErrNotFound))
	}
}

func TestBuntStoreHostList(t *testing.T) {
	assert := assert.New(t)

	store, err := NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 10
	for i := 0; i < size; i++ {
		host := HostFactory.MustCreate().(*Host)
		err := store.StoreHost(host)
		assert.NoError(err)
	}

	hosts, err := store.Hosts()
	assert.NoError(err)
	assert.Equal(10, len(hosts))
}

func TestBuntStoreHostFind(t *testing.T) {
	assert := assert.New(t)

	store, err := NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 20
	for i := 0; i < size; i++ {
		host := HostFactory.MustCreate().(*Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := store.StoreHost(host)
		assert.NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if assert.NoError(err) {
		hosts, err := store.FindHosts(ns)
		assert.NoError(err)
		assert.Equal(10, len(hosts))
	}
}

func TestBuntStoreProvision(t *testing.T) {
	assert := assert.New(t)

	store, err := NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 20
	for i := 0; i < size; i++ {
		host := HostFactory.MustCreate().(*Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := store.StoreHost(host)
		assert.NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if assert.NoError(err) {
		hosts, err := store.FindHosts(ns)
		assert.NoError(err)
		assert.Equal(10, len(hosts))
		for _, host := range hosts {
			assert.False(host.Provision)
		}

		err = store.ProvisionHosts(ns, true)
		assert.NoError(err)

		hosts, err = store.FindHosts(ns)
		assert.NoError(err)
		assert.Equal(10, len(hosts))
		for _, host := range hosts {
			assert.True(host.Provision)
		}
	}
}

func TestBuntStoreSetBootImage(t *testing.T) {
	assert := assert.New(t)

	store, err := NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 20
	for i := 0; i < size; i++ {
		host := HostFactory.MustCreate().(*Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := store.StoreHost(host)
		assert.NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if assert.NoError(err) {
		hosts, err := store.FindHosts(ns)
		assert.NoError(err)
		assert.Equal(10, len(hosts))
		for _, host := range hosts {
			assert.Equal("", host.BootImage)
		}

		err = store.SetBootImage(ns, "centos7")
		assert.NoError(err)

		hosts, err = store.FindHosts(ns)
		assert.NoError(err)
		assert.Equal(10, len(hosts))
		for _, host := range hosts {
			assert.Equal("centos7", host.BootImage)
		}
	}
}

func TestBuntStoreBootImage(t *testing.T) {
	assert := assert.New(t)

	store, err := NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	image := BootImageFactory.MustCreate().(*BootImage)

	err = store.StoreBootImage(image)
	assert.NoError(err)

	testImage, err := store.LoadBootImage(image.Name)
	if assert.NoError(err) {
		assert.Equal(image.Name, testImage.Name)
	}

	badimage := &BootImage{}
	err = store.StoreBootImage(badimage)
	if assert.Error(err) {
		assert.True(errors.Is(err, ErrInvalidData))
	}

	_, err = store.LoadBootImage("notfound")
	if assert.Error(err) {
		assert.True(errors.Is(err, ErrNotFound))
	}

	for i := 0; i < 5; i++ {
		image := BootImageFactory.MustCreate().(*BootImage)
		err := store.StoreBootImage(image)
		assert.NoError(err)
	}

	images, err := store.BootImages()
	if assert.NoError(err) {
		assert.Equal(6, len(images))
	}
}

func BenchmarkBuntStoreWriteHost(b *testing.B) {
	b.StopTimer()
	file := tempfile()
	defer os.Remove(file)

	store, err := NewBuntStore(file)
	defer store.Close()
	if err != nil {
		panic(err)
	}

	size := 5000
	b.StartTimer()
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

func BenchmarkBuntStoreReadAll(b *testing.B) {
	b.StopTimer()
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

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		_, err := store.Hosts()
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkBuntStoreFind(b *testing.B) {
	b.StopTimer()
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
		host.Name = fmt.Sprintf("tux-%04d", i)
		err = store.StoreHost(host)
		if err != nil {
			panic(err)
		}
		hosts[i] = host
	}

	b.SetParallelism(128)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				n := rand.Intn(int(size / 2))
				start := rand.Intn(int(size / 2))
				end := start + n
				if end > size-1 {
					end = size - 1
				}

				n = end - start

				ns, err := nodeset.NewNodeSet(fmt.Sprintf("tux-[%04d-%04d]", start, end))
				if err != nil {
					panic(err)
				}

				hosts, err := store.FindHosts(ns)
				if err != nil {
					panic(err)
				}

				if len(hosts) != n+1 {
					panic("Invalid length")
				}
			}
		})
	}
}

func BenchmarkBuntStoreRandomRead(b *testing.B) {
	b.StopTimer()
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
	b.StartTimer()
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
