package model

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVDB(t *testing.T) {
	dir := tempdir()
	defer os.RemoveAll(dir)

	store, err := NewKVStore(dir)
	defer store.Close()

	assert.Nil(t, err)
}

func TestKVHost(t *testing.T) {
	dir := tempdir()
	defer os.RemoveAll(dir)

	store, err := NewKVStore(dir)
	defer store.Close()

	assert.Nil(t, err)

	mac, err := randmac()
	assert.Nil(t, err)

	bmcMac, err := randmac()
	assert.Nil(t, err)

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
	assert.Nil(t, err)

	testHost, err := store.GetHostByID(host.ID.String())
	assert.Nil(t, err)

	assert.Equal(t, 2, len(testHost.Interfaces))

	nic := testHost.Interface(mac)
	assert.NotNil(t, nic)

	assert.Equal(t, 0, bytes.Compare(nic.MAC, host.Interface(mac).MAC))

	bmc := testHost.InterfaceBMC()
	assert.NotNil(t, bmc)

	assert.Equal(t, 0, bytes.Compare(bmc.MAC, host.InterfaceBMC().MAC))

	assert.Equal(t, nic.FQDN, host.Interface(mac).FQDN)
	assert.Equal(t, bmc.FQDN, host.InterfaceBMC().FQDN)

	hostList, err := store.HostList()
	assert.Nil(t, err)

	assert.Equal(t, 1, len(hostList))

	h2, err := store.GetHostByName(host.Name)
	assert.Nil(t, err)
	assert.Equal(t, host.ID, h2.ID)
}

func TestKVHostList(t *testing.T) {
	dir := tempdir()
	defer os.RemoveAll(dir)

	store, err := NewKVStore(dir)
	defer store.Close()
	assert.Nil(t, err)

	for i := 0; i < 11; i++ {
		mac, err := randmac()
		assert.Nil(t, err)

		bmcMac, err := randmac()
		assert.Nil(t, err)

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
		assert.Nil(t, err)
	}

	hostList, err := store.HostList()
	assert.Nil(t, err)

	assert.Equal(t, 11, len(hostList.FilterPrefix("test-")))
	assert.Equal(t, 2, len(hostList.FilterPrefix("test-1")))
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
