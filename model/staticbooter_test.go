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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/nodeset"
)

func TestStaticBooterLoadJSON(t *testing.T) {
	testHostJSON := strings.NewReader(TestHostListJSON)
	testBootImageJSON := strings.NewReader(TestBootImageListJSON)

	staticBooter := &StaticBooter{}

	err := staticBooter.LoadBootImageJSON(testBootImageJSON)
	assert.Nil(t, err)

	err = staticBooter.LoadHostJSON(testHostJSON)
	assert.Nil(t, err)

	hostList, err := staticBooter.Hosts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hostList))
	assert.Equal(t, "centos6", hostList[0].BootImage)

	ns, err := nodeset.NewNodeSet("tux01")
	assert.Nil(t, err)

	err = staticBooter.SetBootImage(ns, "noexist")
	assert.NotNil(t, err)

	err = staticBooter.SetBootImage(ns, "centos7")
	assert.Nil(t, err)

	hostList, err = staticBooter.Hosts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hostList))
	assert.Equal(t, "centos7", hostList[0].BootImage)
}

const TestHostListJSON = `[
    {
        "firmware": "",
        "id": "1VCnR6qevU5BbihTIvZEhX002CI",
        "interfaces": [
            {
                "bmc": false,
                "fqdn": "tux01.compute.local",
                "ifname": "",
                "ip": "10.10.1.2",
                "mac": "d0:93:ae:e1:b5:2e"
            }
        ],
        "name": "tux01",
        "boot_image": "centos6",
        "provision": true
    }
]`

const TestBootImageListJSON = `[
    {
        "name": "centos7",
        "kernel": "/usr/local/share/image-boot/centos/vmlinuz",
        "initrd": [
            "/usr/local/share/image-boot/centos/ccr-initrd.img"
        ],
        "liveimg": "/usr/local/share/image-boot/node/compute-node.squashfs",
        "install_repo": "centos7"
    }
]`
