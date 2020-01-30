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

package model

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

var TestHostJSON = []byte(`{"firmware": "","id": "1VCnR6qevU5BbihTIvZEhX002CI","interfaces": [{"bmc": false,"fqdn": "tux01.compute.local", "ifname": "", "ip": "10.10.1.2", "mac": "d0:93:ae:e1:b5:2e" } ], "name": "tux01", "boot_image": "centos6", "provision": true }`)

var NetInterfaceFactory = factory.NewFactory(
	&NetInterface{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(2, 10)), nil
}).Attr("FQDN", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
}).Attr("MAC", func(args factory.Args) (interface{}, error) {
	return net.ParseMAC(randomdata.MacAddress())
}).Attr("IP", func(args factory.Args) (interface{}, error) {
	return net.ParseIP(randomdata.IpV4Address()), nil
})

var HostFactory = factory.NewFactory(
	&Host{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
}).OnCreate(func(args factory.Args) error {
	host := args.Instance().(*Host)
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
	&BootImage{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(2, 10)), nil
}).Attr("KernelPath", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
})

func TestFactory(t *testing.T) {
	assert := assert.New(t)

	for i := 0; i < 3; i++ {
		host := HostFactory.MustCreate().(*Host)
		assert.Greater(len(host.Name), 1)
		assert.Equal(2, len(host.Interfaces))
		assert.False(host.ID.IsNil())

		image := BootImageFactory.MustCreate().(*BootImage)
		assert.Greater(len(image.Name), 1)
	}
}

func BenchmarkGJSONUnmarshall(b *testing.B) {
	jsonStr := string(TestHostJSON)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		host := &Host{}
		host.FromJSON(jsonStr)
	}
}

func BenchmarkGJSONMarshall(b *testing.B) {
	jsonStr := string(TestHostJSON)
	host := &Host{}
	host.FromJSON(jsonStr)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		host.ToJSON()
	}
}

func BenchmarkEncodeUnmarshall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var host Host
		err := json.Unmarshal(TestHostJSON, &host)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeMarshall(b *testing.B) {
	var host Host
	err := json.Unmarshal(TestHostJSON, &host)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := json.Marshal(&host)
		if err != nil {
			b.Fatal(err)
		}
	}
}
