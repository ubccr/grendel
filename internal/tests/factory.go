// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tests

import (
	"net"
	"net/netip"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/pkg/model"
)

var TestHostJSON = []byte(`{"firmware": "","uid": "1VCnR6qevU5BbihTIvZEhX002CI","interfaces": [{"bmc": false,"fqdn": "tux01.compute.local", "ifname": "", "ip": "10.10.1.2/24", "mac": "d0:93:ae:e1:b5:2e" } ], "bonds": [{"peers": ["d0:93:ae:e1:b5:2e", "d0:93:ae:e1:b5:2f"], "bmc": false,"fqdn": "tux04.compute.local", "ifname": "bond0", "ip": "10.11.1.2/24", "mac": "" } ], "name": "tux01", "boot_image": "centos6", "provision": true }`)
var TestBootImageJSON = []byte(`{
	"name": "compute",
	"kernel": "/var/grendel/images/centos7/vmlinuz",
	"initrd": [
		"/var/grendel/images/centos7/ccr-initrd.img"
	],
	"liveimg": "/var/grendel/images/compute-node/compute-node-squashfs.img",
	"cmdline": "console=tty0 console=ttyS0 BOOTIF=$mac rd.neednet=1 ip=dhcp ks=$kickstart network ksdevice=bootif ks.device=bootif inst.stage2=$repo/centos7"
}`)

var NetInterfaceFactory = factory.NewFactory(
	&model.NetInterface{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(2, 10)), nil
}).Attr("FQDN", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
}).Attr("MAC", func(args factory.Args) (interface{}, error) {
	return net.ParseMAC(randomdata.MacAddress())
}).Attr("IP", func(args factory.Args) (interface{}, error) {
	return netip.MustParsePrefix(randomdata.IpV4Address() + "/24"), nil
})

var BondFactory = factory.NewFactory(
	&model.Bond{},
).Attr("Peers", func(args factory.Args) (interface{}, error) {
	return []string{randomdata.MacAddress(), randomdata.MacAddress()}, nil
})

var HostFactory = factory.NewFactory(
	&model.Host{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
}).OnCreate(func(args factory.Args) error {
	host := args.Instance().(*model.Host)
	host.Interfaces[0].BMC = false
	host.Interfaces[1].BMC = true
	host.Bonds[0].IP = netip.MustParsePrefix(randomdata.IpV4Address() + "/24")
	host.Bonds[0].FQDN = randomdata.Alphanumeric(randomdata.Number(5, 50))
	host.Bonds[0].Name = "bond0"
	uuid, err := ksuid.NewRandom()
	if err != nil {
		return err
	}
	host.UID = uuid
	return nil
}).SubSliceFactory("Interfaces", NetInterfaceFactory, func() int { return 2 }).SubSliceFactory("Bonds", BondFactory, func() int { return 1 })

var BootImageFactory = factory.NewFactory(
	&model.BootImage{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(2, 10)), nil
}).Attr("KernelPath", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
}).OnCreate(func(args factory.Args) error {
	image := args.Instance().(*model.BootImage)
	image.InitrdPaths = make([]string, 0)
	image.InitrdPaths = append(image.InitrdPaths, randomdata.Alphanumeric(randomdata.Number(5, 50)))
	return nil
})
