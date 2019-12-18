package model

import (
	"bytes"
	"crypto/rand"
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
		MAC:  mac,
		IP:   net.IPv4zero,
		FQDN: "test.localhost",
		BMCAddress: &NetworkAddress{
			MAC:  bmcMac,
			IP:   net.IPv4zero,
			FQDN: "bmc-test.localhost",
		},
	}

	err = store.SaveHost(host)
	if err != nil {
		t.Fatal(err)
	}

	testHost, err := store.GetHost(host.MAC.String())
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(testHost.MAC, host.MAC) != 0 {
		t.Errorf("Incorrect MAC address: got %s should be %s", testHost.MAC, host.MAC)
	}

	if bytes.Compare(testHost.BMCAddress.MAC, host.BMCAddress.MAC) != 0 {
		t.Errorf("Incorrect BMC MAC address: got %s should be %s", testHost.BMCAddress.MAC, host.BMCAddress.MAC)
	}

	if testHost.FQDN != host.FQDN {
		t.Errorf("Incorrect FQDN: got %s should be %s", testHost.FQDN, host.FQDN)
	}

	hostList, err := store.HostList()
	if err != nil {
		t.Fatal(err)
	}

	if len(hostList) != 1 {
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
