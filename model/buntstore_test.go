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
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

type BuntStoreTestSuite struct {
	DataStoreTestSuite
}

func (s *BuntStoreTestSuite) SetupTest() {
	var err error
	ds, err := model.NewBuntStore(":memory:")
	s.Assert().NoError(err)
	s.SetDataStore(ds)
}

func TestBuntStoreTestSuite(t *testing.T) {
	suite.Run(t, new(BuntStoreTestSuite))
}

func tempfile() string {
	name, err := ioutil.TempFile("", "grendel-bunt-")
	if err != nil {
		panic(err)
	}
	return name.Name()
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
		names, err := store.ReverseResolve(pick.Interfaces[0].AddrString())
		if err != nil {
			b.Fatal(err)
		}
		if len(names) != 1 {
			b.Fatalf("names not found")
		}
	}
}
