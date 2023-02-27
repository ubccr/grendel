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

package tests

import (
	"net"
	"net/netip"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/model"
)

var TestHostJSON = []byte(`{"firmware": "","id": "1VCnR6qevU5BbihTIvZEhX002CI","interfaces": [{"bmc": false,"fqdn": "tux01.compute.local", "ifname": "", "ip": "10.10.1.2/24", "mac": "d0:93:ae:e1:b5:2e" } ], "name": "tux01", "boot_image": "centos6", "provision": true }`)
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

var HostFactory = factory.NewFactory(
	&model.Host{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
}).OnCreate(func(args factory.Args) error {
	host := args.Instance().(*model.Host)
	host.Interfaces[0].BMC = false
	host.Interfaces[1].BMC = true
	uuid, err := ksuid.NewRandom()
	if err != nil {
		return err
	}
	host.ID = uuid
	return nil
}).SubSliceFactory("Interfaces", NetInterfaceFactory, func() int { return 2 })

var BootImageFactory = factory.NewFactory(
	&model.BootImage{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(2, 10)), nil
}).Attr("KernelPath", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
})
