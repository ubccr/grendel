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

package model_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/model"
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

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)

	err = store.StoreHost(host)
	assert.NoError(err)

	testHost, err := store.LoadHostFromID(host.ID.String())
	if assert.NoError(err) {
		assert.Equal(2, len(testHost.Interfaces))
	}

	testHost2, err := store.LoadHostFromName(host.Name)
	if assert.NoError(err) {
		assert.Equal(host.Name, testHost2.Name)
		assert.Equal(0, host.Interfaces[0].Addr().Compare(testHost2.Interfaces[0].Addr()))
	}

	testHost3, err := store.LoadHostFromMAC(host.Interfaces[0].MAC.String())
	if assert.NoError(err) {
		assert.Equal(host.Name, testHost3.Name)
		assert.Equal(host.Interfaces[0].MAC.String(), testHost3.Interfaces[0].MAC.String())
	}

	testIPs, err := store.ResolveIPv4(host.Interfaces[0].FQDN)
	if assert.NoError(err) {
		if assert.Equal(1, len(testIPs)) {
			assert.Equal(host.Interfaces[0].AddrString(), testIPs[0].String())
		}
	}

	testNames, err := store.ReverseResolve(host.Interfaces[0].AddrString())
	if assert.NoError(err) {
		if assert.Equal(1, len(testNames)) {
			assert.Equal(host.Interfaces[0].FQDN, testNames[0])
		}
	}

	badhost := &model.Host{}
	err = store.StoreHost(badhost)
	if assert.Error(err) {
		assert.True(errors.Is(err, model.ErrInvalidData))
	}

	_, err = store.LoadHostFromID("notfound")
	if assert.Error(err) {
		assert.True(errors.Is(err, model.ErrNotFound))
	}

	_, err = store.LoadHostFromName("notfound")
	if assert.Error(err) {
		assert.True(errors.Is(err, model.ErrNotFound))
	}

	_, err = store.LoadHostFromMAC("notfound")
	if assert.Error(err) {
		assert.True(errors.Is(err, model.ErrNotFound))
	}
}

func TestBuntStoreIfname(t *testing.T) {
	assert := assert.New(t)

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)

	err = store.StoreHost(host)
	assert.NoError(err)

	testHost, err := store.LoadHostFromName(host.Name)
	if assert.NoError(err) {
		assert.Equal(host.Interfaces[0].Name, testHost.Interfaces[0].Name)
	}
}

func TestBuntStoreHostList(t *testing.T) {
	assert := assert.New(t)

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 10
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		err := store.StoreHost(host)
		assert.NoError(err)
	}

	hosts, err := store.Hosts()
	assert.NoError(err)
	assert.Equal(10, len(hosts))
}

func TestBuntStoreHostFind(t *testing.T) {
	assert := assert.New(t)

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
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

func TestBuntStoreFindTags(t *testing.T) {
	assert := assert.New(t)

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 10
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		if (i % 2) == 0 {
			host.Tags = []string{"k11", "wanda"}
		} else if (i % 2) != 0 {
			host.Tags = []string{"k16", "vision"}
		}
		err := store.StoreHost(host)
		assert.NoError(err)
	}

	ns, err := store.FindTags([]string{"k16"})
	if assert.NoError(err) {
		assert.Equal(5, ns.Len())
	}

	ns, err = store.FindTags([]string{"vision"})
	if assert.NoError(err) {
		assert.Equal(5, ns.Len())
	}

	ns, err = store.FindTags([]string{"vision", "k11"})
	if assert.NoError(err) {
		assert.Equal(10, ns.Len())
	}

	ns, err = store.FindTags([]string{"harkness", "rambeau"})
	if assert.Error(err) {
		assert.True(errors.Is(err, model.ErrNotFound))
	}

	ns, err = nodeset.NewNodeSet("tux-[05-08]")
	if assert.NoError(err) {
		err := store.TagHosts(ns, []string{"harkness"})
		assert.NoError(err)
	}

	ns, err = store.FindTags([]string{"harkness"})
	if assert.NoError(err) {
		assert.Equal(4, ns.Len())
	}

	ns, err = nodeset.NewNodeSet("tux-[00-10]")
	if assert.NoError(err) {
		err := store.UntagHosts(ns, []string{"vision"})
		assert.NoError(err)
	}

	ns, err = store.FindTags([]string{"vision"})
	if assert.Error(err) {
		assert.True(errors.Is(err, model.ErrNotFound))
	}
}

func TestBuntStoreProvision(t *testing.T) {
	assert := assert.New(t)

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
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

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
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

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)

	image.ProvisionTemplates = map[string]string{
		"kickstart":    "kickstart.tmpl",
		"post-install": "post-install.tmpl",
	}

	err = store.StoreBootImage(image)
	assert.NoError(err)

	testImage, err := store.LoadBootImage(image.Name)
	if assert.NoError(err) {
		assert.Equal(image.Name, testImage.Name)
		assert.Contains(testImage.ProvisionTemplates, "post-install")
		assert.Contains(testImage.ProvisionTemplates, "kickstart")
	}

	badimage := &model.BootImage{}
	err = store.StoreBootImage(badimage)
	if assert.Error(err) {
		assert.True(errors.Is(err, model.ErrInvalidData))
	}

	_, err = store.LoadBootImage("notfound")
	if assert.Error(err) {
		assert.True(errors.Is(err, model.ErrNotFound))
	}

	for i := 0; i < 5; i++ {
		image := tests.BootImageFactory.MustCreate().(*model.BootImage)
		err := store.StoreBootImage(image)
		assert.NoError(err)
	}

	images, err := store.BootImages()
	if assert.NoError(err) {
		assert.Equal(6, len(images))
	}
}

func TestBuntStoreBootImageDelete(t *testing.T) {
	assert := assert.New(t)

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)

	err = store.StoreBootImage(image)
	assert.NoError(err)

	testImage, err := store.LoadBootImage(image.Name)
	if assert.NoError(err) {
		assert.Equal(image.Name, testImage.Name)
	}

	err = store.DeleteBootImages([]string{testImage.Name})
	if assert.NoError(err) {
		_, err = store.LoadBootImage(testImage.Name)
		if assert.Error(err) {
			assert.True(errors.Is(err, model.ErrNotFound))
		}
	}
}

func TestBuntStoreUpdate(t *testing.T) {
	assert := assert.New(t)

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)

	err = store.StoreHost(host)
	assert.NoError(err)

	testHost, err := store.LoadHostFromID(host.ID.String())
	if assert.NoError(err) {
		assert.Equal(2, len(testHost.Interfaces))
	}

	// Store host with same name is update
	hostDup := tests.HostFactory.MustCreate().(*model.Host)
	hostDup.ID = host.ID
	hostDup.Name = host.Name
	err = store.StoreHost(hostDup)
	if assert.NoError(err) {
		hosts, err := store.Hosts()
		assert.NoError(err)
		assert.Equal(1, len(hosts))
	}

	// Store host with different name gets new ID
	hostDup = tests.HostFactory.MustCreate().(*model.Host)
	hostDup.ID = host.ID
	hostDup.Name = "cpn-new"
	err = store.StoreHost(hostDup)
	if assert.NoError(err) {
		hosts, err := store.Hosts()
		assert.NoError(err)
		assert.Equal(2, len(hosts))
		idCheck := ""
		for _, h := range hosts {
			assert.NotEqual(idCheck, h.ID.String())
			idCheck = h.ID.String()
		}
	}
}

func TestBuntStoreHostDelete(t *testing.T) {
	assert := assert.New(t)

	store, err := model.NewBuntStore(":memory:")
	defer store.Close()
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)

	err = store.StoreHost(host)
	assert.NoError(err)

	testHost, err := store.LoadHostFromID(host.ID.String())
	if assert.NoError(err) {
		assert.Equal(2, len(testHost.Interfaces))
	}

	ns, err := nodeset.NewNodeSet(testHost.Name)
	if assert.NoError(err) {
		err := store.DeleteHosts(ns)
		assert.NoError(err)

		_, err = store.LoadHostFromID(host.ID.String())
		if assert.Error(err) {
			assert.True(errors.Is(err, model.ErrNotFound))
		}
	}
}

func BenchmarkBuntStoreWriteHosts(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := model.NewBuntStore(file)
	defer store.Close()
	if err != nil {
		b.Fatal(err)
	}

	size := 5000
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := store.StoreHosts(hosts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuntStoreWriteSingleHost(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := model.NewBuntStore(file)
	defer store.Close()
	if err != nil {
		b.Fatal(err)
	}

	size := 5000
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < size; i++ {
			err := store.StoreHost(hosts[i])
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkBuntStoreReadAll(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := model.NewBuntStore(file)
	defer store.Close()
	if err != nil {
		b.Fatal(err)
	}

	size := 5000
	rand.Seed(time.Now().UnixNano())
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	err = store.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		list, err := store.Hosts()
		if err != nil {
			b.Fatal(err)
		}
		if len(list) != size {
			b.Fatalf("wrong size: expected %d got %d", size, len(list))
		}
	}
}

func BenchmarkBuntStoreParallelFind(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := model.NewBuntStore(file)
	defer store.Close()
	if err != nil {
		b.Fatal(err)
	}

	size := 5000
	rand.Seed(time.Now().UnixNano())
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%04d", i)
		hosts[i] = host
	}

	err = store.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.SetParallelism(128)
	b.ResetTimer()
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
					b.Fatal(err)
				}

				hosts, err := store.FindHosts(ns)
				if err != nil {
					b.Fatal(err)
				}

				if len(hosts) != n+1 {
					b.Fatalf("wrong host count found: expected %d got %d", n+1, len(hosts))
				}
			}
		})
	}
}

func BenchmarkBuntStoreRandomParallelReads(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := model.NewBuntStore(file)
	defer store.Close()
	if err != nil {
		b.Fatal(err)
	}

	size := 5000
	rand.Seed(time.Now().UnixNano())
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	err = store.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.SetParallelism(128)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				pick := hosts[rand.Intn(size)]
				_, err := store.LoadHostFromID(pick.ID.String())
				if err != nil {
					b.Fatal(err)
				}

				_, err = store.LoadHostFromName(pick.Name)
				if err != nil {
					b.Fatal(err)
				}

				ips, err := store.ResolveIPv4(pick.Interfaces[0].FQDN)
				if err != nil || len(ips) != 1 {
					b.Fatal(err)
				}
			}
		})

	}
}

func BenchmarkBuntStoreResolveIPv4(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := model.NewBuntStore(file)
	defer store.Close()
	if err != nil {
		b.Fatal(err)
	}

	size := 5000
	rand.Seed(time.Now().UnixNano())
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	err = store.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pick := hosts[rand.Intn(size)]
		ips, err := store.ResolveIPv4(pick.Interfaces[0].FQDN)
		if err != nil {
			b.Fatal(err)
		}
		if len(ips) != 1 {
			b.Fatalf("IPs not found")
		}
	}
}

func BenchmarkBuntStoreReverseResolve(b *testing.B) {
	file := tempfile()
	defer os.Remove(file)

	store, err := model.NewBuntStore(file)
	defer store.Close()
	if err != nil {
		b.Fatal(err)
	}

	size := 5000
	rand.Seed(time.Now().UnixNano())
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	err = store.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pick := hosts[rand.Intn(size)]
		names, err := store.ReverseResolve(pick.Interfaces[0].IP.String())
		if err != nil {
			b.Fatal(err)
		}
		if len(names) != 1 {
			b.Fatalf("names not found")
		}
	}
}
