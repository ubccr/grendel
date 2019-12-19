package model

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func TestKVDB(t *testing.T) {
	dir := tempdir()
	defer os.RemoveAll(dir)

	store, err := NewKVStore(dir)
	defer store.Close()

	if err != nil {
		t.Fatal(err)
	}
}

func TestKVHost(t *testing.T) {
	dir := tempdir()
	defer os.RemoveAll(dir)

	store, err := NewKVStore(dir)
	defer store.Close()

	if err != nil {
		t.Fatal(err)
	}

	mac, err := randmac()
	if err != nil {
		t.Fatal(err)
	}

	bmcMac, err := randmac()
	if err != nil {
		t.Fatal(err)
	}

	host := &Host{
		Interfaces: []*NetInterface{
			&NetInterface{
				MAC:  mac,
				IP:   net.IPv4zero,
				FQDN: "test.localhost",
			},
			&NetInterface{
				MAC:  bmcMac,
				IP:   net.IPv4zero,
				FQDN: "bmc-test.localhost",
				BMC:  true,
			},
		},
	}

	err = store.SaveHost(host)
	if err != nil {
		t.Fatal(err)
	}

	testHost, err := store.GetHost(host.ID.String())
	if err != nil {
		t.Fatal(err)
	}

	if len(testHost.Interfaces) != 2 {
		t.Errorf("Host should have 2 network addresses got: %d", len(testHost.Interfaces))
	}

	nic := testHost.Interface(mac)
	if nic == nil {
		t.Errorf("Failed to find network interface for host")
	}

	if bytes.Compare(nic.MAC, host.Interface(mac).MAC) != 0 {
		t.Errorf("Incorrect MAC address: got %s should be %s", nic.MAC, host.Interface(mac).MAC)
	}

	bmc := testHost.InterfaceBMC()
	if nic == nil {
		t.Errorf("Failed to find BMC interface for host")
	}

	if bytes.Compare(bmc.MAC, host.InterfaceBMC().MAC) != 0 {
		t.Errorf("Incorrect BMC MAC address: got %s should be %s", bmc.MAC, host.InterfaceBMC().MAC)
	}

	if nic.FQDN != host.Interface(mac).FQDN {
		t.Errorf("Incorrect FQDN: got %s should be %s", nic.FQDN, host.Interface(mac).FQDN)
	}

	if bmc.FQDN != host.InterfaceBMC().FQDN {
		t.Errorf("Incorrect FQDN: got %s should be %s", bmc.FQDN, host.InterfaceBMC().FQDN)
	}

	hostList, err := store.HostList()
	if err != nil {
		t.Fatal(err)
	}

	if len(hostList) != 1 {
		t.Errorf("Incorrect size of host list: got %d should be %d", len(hostList), 1)
	}
}

func TestKVHostList(t *testing.T) {
	dir := tempdir()
	defer os.RemoveAll(dir)

	store, err := NewKVStore(dir)
	defer store.Close()

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 11; i++ {
		mac, err := randmac()
		if err != nil {
			t.Fatal(err)
		}

		bmcMac, err := randmac()
		if err != nil {
			t.Fatal(err)
		}

		host := &Host{
			Name: fmt.Sprintf("test-%d", i),
			Interfaces: []*NetInterface{
				&NetInterface{
					MAC:  mac,
					IP:   net.IPv4zero,
					FQDN: fmt.Sprintf("test-%d.localhost", i),
				},
				&NetInterface{
					MAC:  bmcMac,
					IP:   net.IPv4zero,
					FQDN: fmt.Sprintf("bmc-test-%d.localhost", i),
					BMC:  true,
				},
			},
		}

		err = store.SaveHost(host)
		if err != nil {
			t.Fatalf("Failed to save host: %#v - %s", host, err)
		}
	}

	hostList, err := store.HostList()
	if err != nil {
		t.Fatal(err)
	}

	if len(hostList.FilterPrefix("test-")) != 11 {
		t.Errorf("Incorrect size of host list: got %d should be %d", len(hostList), 1)
	}

	if len(hostList.FilterPrefix("test-1")) != 2 {
		t.Errorf("Incorrect size of host list: got %d should be %d", len(hostList), 1)
	}
}

func randmac() (net.HardwareAddr, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	buf[0] |= 2
	return net.HardwareAddr(buf), nil
}

func tempdir() string {
	name, err := ioutil.TempDir("", "grendel-")
	if err != nil {
		panic(err)
	}
	return name
}
