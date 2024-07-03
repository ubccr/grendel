package model_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/model"
)

func BenchmarkRqliteWriteHosts(b *testing.B) {
	store, err := model.NewRqliteStore("http://localhost:4001")
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

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

func BenchmarkRqliteWriteSingleHost(b *testing.B) {
	store, err := model.NewRqliteStore("http://localhost:4001")
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

func BenchmarkRqliteReadAll(b *testing.B) {
	store, err := model.NewRqliteStore("http://localhost:4001")
	defer store.DeleteHosts()
	if err != nil {
		b.Fatal(err)
	}

	size := 2
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
