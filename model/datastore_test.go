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
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

type DataStoreTestSuite struct {
	suite.Suite
	ds model.DataStore
}

func (suite *DataStoreTestSuite) SetDataStore(ds model.DataStore) {
	suite.ds = ds
}

func (suite *DataStoreTestSuite) TestUser() {
	adminUsername := "admin"
	adminPassword := "SuperSecureAdminPassword1234!@#$"
	userUsername := "user"
	userPassword := "1234"

	defer suite.ds.Close()

	role, err := suite.ds.StoreUser(adminUsername, adminPassword)
	suite.Assert().NoError(err)
	suite.Assert().Equal(role, "admin")
	role, err = suite.ds.StoreUser(userUsername, userPassword)
	suite.Assert().NoError(err)
	suite.Assert().Equal(role, "disabled")

	authenticated, role, err := suite.ds.VerifyUser("admin", adminPassword)
	suite.Assert().NoError(err)
	suite.Assert().Equal(role, "admin")
	suite.Assert().Equal(authenticated, true)
	authenticated, role, err = suite.ds.VerifyUser("user", userPassword)
	suite.Assert().NoError(err)
	suite.Assert().Equal(role, "disabled")
	suite.Assert().Equal(authenticated, true)

	users, err := suite.ds.GetUsers()
	suite.Assert().NoError(err)
	suite.Assert().Equal(users[0].Username, "admin")
	suite.Assert().Contains(users[1].Username, "user")

	err = suite.ds.UpdateUser("user", "user")
	suite.Assert().NoError(err)
	authenticated, role, err = suite.ds.VerifyUser("user", userPassword)
	suite.Assert().NoError(err)
	suite.Assert().Equal(role, "user")
	suite.Assert().Equal(authenticated, true)

	err = suite.ds.DeleteUser("user")
	suite.Assert().NoError(err)
	users, err = suite.ds.GetUsers()
	suite.Assert().NoError(err)
	suite.Assert().Equal(len(users), 1)
}

func (suite *DataStoreTestSuite) TestHost() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := suite.ds.StoreHost(host)
	suite.Assert().NoError(err)

	testHost, err := suite.ds.LoadHostFromID(host.ID.String())
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(2, len(testHost.Interfaces))
	}

	testHost2, err := suite.ds.LoadHostFromName(host.Name)
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(host.Name, testHost2.Name)
		suite.Assert().Equal(0, host.Interfaces[0].Addr().Compare(testHost2.Interfaces[0].Addr()))
	}

	testHost3, err := suite.ds.LoadHostFromMAC(host.Interfaces[0].MAC.String())
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(host.Name, testHost3.Name)
		suite.Assert().Equal(host.Interfaces[0].MAC.String(), testHost3.Interfaces[0].MAC.String())
	}

	testIPs, err := suite.ds.ResolveIPv4(host.Interfaces[0].FQDN)
	if suite.Assert().NoError(err) {
		if suite.Assert().Equal(1, len(testIPs)) {
			suite.Assert().Equal(host.Interfaces[0].AddrString(), testIPs[0].String())
		}
	}

	testBondIPs, err := suite.ds.ResolveIPv4(host.Bonds[0].FQDN)
	if suite.Assert().NoError(err) {
		if suite.Assert().Equal(1, len(testBondIPs)) {
			suite.Assert().Equal(host.Bonds[0].AddrString(), testBondIPs[0].String())
		}
	}

	testNames, err := suite.ds.ReverseResolve(host.Interfaces[0].AddrString())
	if suite.Assert().NoError(err) {
		if suite.Assert().Equal(1, len(testNames)) {
			suite.Assert().Equal(host.Interfaces[0].FQDN, testNames[0])
		}
	}

	testBondNames, err := suite.ds.ReverseResolve(host.Bonds[0].AddrString())
	if suite.Assert().NoError(err) {
		if suite.Assert().Equal(1, len(testBondNames)) {
			suite.Assert().Equal(host.Bonds[0].FQDN, testBondNames[0])
		}
	}

	badhost := &model.Host{}
	err = suite.ds.StoreHost(badhost)
	if suite.Assert().Error(err) {
		suite.Assert().True(errors.Is(err, model.ErrInvalidData))
	}

	_, err = suite.ds.LoadHostFromID("notfound")
	if suite.Assert().Error(err) {
		suite.Assert().True(errors.Is(err, model.ErrNotFound))
	}

	_, err = suite.ds.LoadHostFromName("notfound")
	if suite.Assert().Error(err) {
		suite.Assert().True(errors.Is(err, model.ErrNotFound))
	}

	_, err = suite.ds.LoadHostFromMAC("notfound")
	if suite.Assert().Error(err) {
		suite.Assert().True(errors.Is(err, model.ErrNotFound))
	}
}

func (suite *DataStoreTestSuite) TestIfname() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := suite.ds.StoreHost(host)
	suite.Assert().NoError(err)

	testHost, err := suite.ds.LoadHostFromName(host.Name)
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(host.Interfaces[0].Name, testHost.Interfaces[0].Name)
	}
}

func (suite *DataStoreTestSuite) TestBonds() {
	host := tests.HostFactory.MustCreate().(*model.Host)
	host.Bonds[0].Peers[0] = host.Interfaces[0].MAC.String()
	host.Bonds[0].Peers[1] = host.Interfaces[1].Name

	err := suite.ds.StoreHost(host)
	suite.Assert().NoError(err)

	testHost, err := suite.ds.LoadHostFromName(host.Name)
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(host.Bonds[0].Peers, testHost.Bonds[0].Peers)
		suite.Assert().Equal(host.Bonds[0].IP, testHost.Bonds[0].IP)
		suite.Assert().Equal(host.Bonds[0].Name, testHost.Bonds[0].Name)
		suite.Assert().True(host.InterfaceBonded(host.Interfaces[0].MAC.String()))
		suite.Assert().False(host.InterfaceBonded(host.Interfaces[1].MAC.String()))
		suite.Assert().True(host.InterfaceBonded(host.Interfaces[1].Name))
	}
}

func (suite *DataStoreTestSuite) TestHostList() {
	size := 10
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		err := suite.ds.StoreHost(host)
		suite.Assert().NoError(err)
	}

	hosts, err := suite.ds.Hosts()
	suite.Assert().NoError(err)
	suite.Assert().Equal(10, len(hosts))
}

func (suite *DataStoreTestSuite) TestHostFind() {
	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := suite.ds.StoreHost(host)
		suite.Assert().NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if suite.Assert().NoError(err) {
		hosts, err := suite.ds.FindHosts(ns)
		suite.Assert().NoError(err)
		suite.Assert().Equal(10, len(hosts))
	}
}

func (suite *DataStoreTestSuite) TestFindTags() {
	size := 10
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		if (i % 2) == 0 {
			host.Tags = []string{"k11", "wanda"}
		} else if (i % 2) != 0 {
			host.Tags = []string{"k16", "vision"}
		}
		err := suite.ds.StoreHost(host)
		suite.Assert().NoError(err)
	}

	ns, err := suite.ds.FindTags([]string{"k16"})
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(5, ns.Len())
	}

	ns, err = suite.ds.FindTags([]string{"vision"})
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(5, ns.Len())
	}

	ns, err = suite.ds.FindTags([]string{"vision", "k11"})
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(10, ns.Len())
	}

	ns, err = suite.ds.FindTags([]string{"harkness", "rambeau"})
	if suite.Assert().Error(err) {
		suite.Assert().True(errors.Is(err, model.ErrNotFound))
	}

	ns, err = nodeset.NewNodeSet("tux-[05-08]")
	if suite.Assert().NoError(err) {
		err := suite.ds.TagHosts(ns, []string{"harkness"})
		suite.Assert().NoError(err)
	}

	ns, err = suite.ds.FindTags([]string{"harkness"})
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(4, ns.Len())
	}

	ns, err = nodeset.NewNodeSet("tux-[00-10]")
	if suite.Assert().NoError(err) {
		err := suite.ds.UntagHosts(ns, []string{"vision"})
		suite.Assert().NoError(err)
	}

	ns, err = suite.ds.FindTags([]string{"vision"})
	if suite.Assert().Error(err) {
		suite.Assert().True(errors.Is(err, model.ErrNotFound))
	}
}

func (suite *DataStoreTestSuite) TestProvision() {
	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := suite.ds.StoreHost(host)
		suite.Assert().NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if suite.Assert().NoError(err) {
		hosts, err := suite.ds.FindHosts(ns)
		suite.Assert().NoError(err)
		suite.Assert().Equal(10, len(hosts))
		for _, host := range hosts {
			suite.Assert().False(host.Provision)
		}

		err = suite.ds.ProvisionHosts(ns, true)
		suite.Assert().NoError(err)

		hosts, err = suite.ds.FindHosts(ns)
		suite.Assert().NoError(err)
		suite.Assert().Equal(10, len(hosts))
		for _, host := range hosts {
			suite.Assert().True(host.Provision)
		}
	}
}

func (suite *DataStoreTestSuite) TestSetBootImage() {
	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := suite.ds.StoreHost(host)
		suite.Assert().NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if suite.Assert().NoError(err) {
		hosts, err := suite.ds.FindHosts(ns)
		suite.Assert().NoError(err)
		suite.Assert().Equal(10, len(hosts))
		for _, host := range hosts {
			suite.Assert().Equal("", host.BootImage)
		}

		err = suite.ds.SetBootImage(ns, "centos7")
		suite.Assert().NoError(err)

		hosts, err = suite.ds.FindHosts(ns)
		suite.Assert().NoError(err)
		suite.Assert().Equal(10, len(hosts))
		for _, host := range hosts {
			suite.Assert().Equal("centos7", host.BootImage)
		}
	}
}

func (suite *DataStoreTestSuite) TestBootImage() {
	image := tests.BootImageFactory.MustCreate().(*model.BootImage)

	image.ProvisionTemplates = map[string]string{
		"kickstart":    "kickstart.tmpl",
		"post-install": "post-install.tmpl",
	}

	err := suite.ds.StoreBootImage(image)
	suite.Assert().NoError(err)

	testImage, err := suite.ds.LoadBootImage(image.Name)
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(image.Name, testImage.Name)
		suite.Assert().Contains(testImage.ProvisionTemplates, "post-install")
		suite.Assert().Contains(testImage.ProvisionTemplates, "kickstart")
	}

	badimage := &model.BootImage{}
	err = suite.ds.StoreBootImage(badimage)
	if suite.Assert().Error(err) {
		suite.Assert().True(errors.Is(err, model.ErrInvalidData))
	}

	_, err = suite.ds.LoadBootImage("notfound")
	if suite.Assert().Error(err) {
		suite.Assert().True(errors.Is(err, model.ErrNotFound))
	}

	for i := 0; i < 5; i++ {
		image := tests.BootImageFactory.MustCreate().(*model.BootImage)
		err := suite.ds.StoreBootImage(image)
		suite.Assert().NoError(err)
	}

	images, err := suite.ds.BootImages()
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(6, len(images))
	}
}

func (suite *DataStoreTestSuite) TestBootImageDelete() {
	image := tests.BootImageFactory.MustCreate().(*model.BootImage)

	err := suite.ds.StoreBootImage(image)
	suite.Assert().NoError(err)

	testImage, err := suite.ds.LoadBootImage(image.Name)
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(image.Name, testImage.Name)
	}

	err = suite.ds.DeleteBootImages([]string{testImage.Name})
	if suite.Assert().NoError(err) {
		_, err = suite.ds.LoadBootImage(testImage.Name)
		if suite.Assert().Error(err) {
			suite.Assert().True(errors.Is(err, model.ErrNotFound))
		}
	}
}

func (suite *DataStoreTestSuite) TestHostUpdate() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := suite.ds.StoreHost(host)
	suite.Assert().NoError(err)

	testHost, err := suite.ds.LoadHostFromID(host.ID.String())
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(2, len(testHost.Interfaces))
	}

	// Store host with same name is update
	hostDup := tests.HostFactory.MustCreate().(*model.Host)
	hostDup.ID = host.ID
	hostDup.Name = host.Name
	err = suite.ds.StoreHost(hostDup)
	if suite.Assert().NoError(err) {
		hosts, err := suite.ds.Hosts()
		suite.Assert().NoError(err)
		suite.Assert().Equal(1, len(hosts))
	}

	// Store host with different name gets new ID
	hostDup = tests.HostFactory.MustCreate().(*model.Host)
	hostDup.ID = host.ID
	hostDup.Name = "cpn-new"
	err = suite.ds.StoreHost(hostDup)
	if suite.Assert().NoError(err) {
		hosts, err := suite.ds.Hosts()
		suite.Assert().NoError(err)
		suite.Assert().Equal(2, len(hosts))
		idCheck := ""
		for _, h := range hosts {
			suite.Assert().NotEqual(idCheck, h.ID.String())
			idCheck = h.ID.String()
		}
	}
}

func (suite *DataStoreTestSuite) TestHostDelete() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := suite.ds.StoreHost(host)
	suite.Assert().NoError(err)

	testHost, err := suite.ds.LoadHostFromID(host.ID.String())
	if suite.Assert().NoError(err) {
		suite.Assert().Equal(2, len(testHost.Interfaces))
	}

	ns, err := nodeset.NewNodeSet(testHost.Name)
	if suite.Assert().NoError(err) {
		err := suite.ds.DeleteHosts(ns)
		suite.Assert().NoError(err)

		_, err = suite.ds.LoadHostFromID(host.ID.String())
		if suite.Assert().Error(err) {
			suite.Assert().True(errors.Is(err, model.ErrNotFound))
		}
	}
}

func BenchmarkGJSONUnmarshall(b *testing.B) {
	jsonStr := string(tests.TestHostJSON)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		host := &model.Host{}
		host.FromJSON(jsonStr)
	}
}

func BenchmarkGJSONMarshall(b *testing.B) {
	jsonStr := string(tests.TestHostJSON)
	host := &model.Host{}
	host.FromJSON(jsonStr)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		host.ToJSON()
	}
}

func BenchmarkEncodeUnmarshall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var host model.Host
		err := json.Unmarshal(tests.TestHostJSON, &host)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeMarshall(b *testing.B) {
	var host model.Host
	err := json.Unmarshal(tests.TestHostJSON, &host)
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
