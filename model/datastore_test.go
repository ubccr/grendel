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
	"errors"
	"fmt"

	"github.com/stretchr/testify/suite"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

type DataStoreTestSuite struct {
	suite.Suite
	db model.DataStore
}

func (s *DataStoreTestSuite) SetDataStore(db model.DataStore) {
	s.db = db
}

func (s *DataStoreTestSuite) TestUser() {
	adminUsername := "admin"
	adminPassword := "SuperSecureAdminPassword1234!@#$"
	userUsername := "user"
	userPassword := "1234"

	defer s.db.Close()

	role, err := s.db.StoreUser(adminUsername, adminPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(role, "admin")
	role, err = s.db.StoreUser(userUsername, userPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(role, "disabled")

	authenticated, role, err := s.db.VerifyUser("admin", adminPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(role, "admin")
	s.Assert().Equal(authenticated, true)
	authenticated, role, err = s.db.VerifyUser("user", userPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(role, "disabled")
	s.Assert().Equal(authenticated, true)

	users, err := s.db.GetUsers()
	s.Assert().NoError(err)
	s.Assert().Equal(users[0].Username, "admin")
	s.Assert().Contains(users[1].Username, "user")

	err = s.db.UpdateUser("user", "user")
	s.Assert().NoError(err)
	authenticated, role, err = s.db.VerifyUser("user", userPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(role, "user")
	s.Assert().Equal(authenticated, true)

	err = s.db.DeleteUser("user")
	s.Assert().NoError(err)
	users, err = s.db.GetUsers()
	s.Assert().NoError(err)
	s.Assert().Equal(len(users), 1)
}

func (s *DataStoreTestSuite) TestHost() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromID(host.ID.String())
	if s.Assert().NoError(err) {
		s.Assert().Equal(2, len(testHost.Interfaces))
	}

	testHost2, err := s.db.LoadHostFromName(host.Name)
	if s.Assert().NoError(err) {
		s.Assert().Equal(host.Name, testHost2.Name)
		s.Assert().Equal(0, host.Interfaces[0].Addr().Compare(testHost2.Interfaces[0].Addr()))
	}

	testHost3, err := s.db.LoadHostFromMAC(host.Interfaces[0].MAC.String())
	if s.Assert().NoError(err) {
		s.Assert().Equal(host.Name, testHost3.Name)
		s.Assert().Equal(host.Interfaces[0].MAC.String(), testHost3.Interfaces[0].MAC.String())
	}

	testIPs, err := s.db.ResolveIPv4(host.Interfaces[0].FQDN)
	if s.Assert().NoError(err) {
		if s.Assert().Equal(1, len(testIPs)) {
			s.Assert().Equal(host.Interfaces[0].AddrString(), testIPs[0].String())
		}
	}

	testBondIPs, err := s.db.ResolveIPv4(host.Bonds[0].FQDN)
	if s.Assert().NoError(err) {
		if s.Assert().Equal(1, len(testBondIPs)) {
			s.Assert().Equal(host.Bonds[0].AddrString(), testBondIPs[0].String())
		}
	}

	testNames, err := s.db.ReverseResolve(host.Interfaces[0].AddrString())
	if s.Assert().NoError(err) {
		if s.Assert().Equal(1, len(testNames)) {
			s.Assert().Equal(host.Interfaces[0].FQDN, testNames[0])
		}
	}

	testBondNames, err := s.db.ReverseResolve(host.Bonds[0].AddrString())
	if s.Assert().NoError(err) {
		if s.Assert().Equal(1, len(testBondNames)) {
			s.Assert().Equal(host.Bonds[0].FQDN, testBondNames[0])
		}
	}

	badhost := &model.Host{}
	err = s.db.StoreHost(badhost)
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, model.ErrInvalidData))
	}

	_, err = s.db.LoadHostFromID("notfound")
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, model.ErrNotFound))
	}

	_, err = s.db.LoadHostFromName("notfound")
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, model.ErrNotFound))
	}

	_, err = s.db.LoadHostFromMAC("notfound")
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, model.ErrNotFound))
	}
}

func (s *DataStoreTestSuite) TestIfname() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromName(host.Name)
	if s.Assert().NoError(err) {
		s.Assert().Equal(host.Interfaces[0].Name, testHost.Interfaces[0].Name)
	}
}

func (s *DataStoreTestSuite) TestBonds() {
	host := tests.HostFactory.MustCreate().(*model.Host)
	host.Bonds[0].Peers[0] = host.Interfaces[0].MAC.String()
	host.Bonds[0].Peers[1] = host.Interfaces[1].Name

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromName(host.Name)
	if s.Assert().NoError(err) {
		s.Assert().Equal(host.Bonds[0].Peers, testHost.Bonds[0].Peers)
		s.Assert().Equal(host.Bonds[0].IP, testHost.Bonds[0].IP)
		s.Assert().Equal(host.Bonds[0].Name, testHost.Bonds[0].Name)
		s.Assert().True(host.InterfaceBonded(host.Interfaces[0].MAC.String()))
		s.Assert().False(host.InterfaceBonded(host.Interfaces[1].MAC.String()))
		s.Assert().True(host.InterfaceBonded(host.Interfaces[1].Name))
	}
}

func (s *DataStoreTestSuite) TestHostList() {
	size := 10
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		err := s.db.StoreHost(host)
		s.Assert().NoError(err)
	}

	hosts, err := s.db.Hosts()
	s.Assert().NoError(err)
	s.Assert().Equal(10, len(hosts))
}

func (s *DataStoreTestSuite) TestHostFind() {
	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := s.db.StoreHost(host)
		s.Assert().NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if s.Assert().NoError(err) {
		hosts, err := s.db.FindHosts(ns)
		s.Assert().NoError(err)
		s.Assert().Equal(10, len(hosts))
	}
}

func (s *DataStoreTestSuite) TestFindTags() {
	size := 10
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		if (i % 2) == 0 {
			host.Tags = []string{"k11", "wanda"}
		} else if (i % 2) != 0 {
			host.Tags = []string{"k16", "vision"}
		}
		err := s.db.StoreHost(host)
		s.Assert().NoError(err)
	}

	ns, err := s.db.FindTags([]string{"k16"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(5, ns.Len())
	}

	ns, err = s.db.FindTags([]string{"vision"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(5, ns.Len())
	}

	ns, err = s.db.FindTags([]string{"vision", "k11"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(10, ns.Len())
	}

	ns, err = s.db.FindTags([]string{"harkness", "rambeau"})
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, model.ErrNotFound))
	}

	ns, err = nodeset.NewNodeSet("tux-[05-08]")
	if s.Assert().NoError(err) {
		err := s.db.TagHosts(ns, []string{"harkness"})
		s.Assert().NoError(err)
	}

	ns, err = s.db.FindTags([]string{"harkness"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(4, ns.Len())
	}

	ns, err = nodeset.NewNodeSet("tux-[00-10]")
	if s.Assert().NoError(err) {
		err := s.db.UntagHosts(ns, []string{"vision"})
		s.Assert().NoError(err)
	}

	ns, err = s.db.FindTags([]string{"vision"})
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, model.ErrNotFound))
	}
}

func (s *DataStoreTestSuite) TestProvision() {
	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := s.db.StoreHost(host)
		s.Assert().NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if s.Assert().NoError(err) {
		hosts, err := s.db.FindHosts(ns)
		s.Assert().NoError(err)
		s.Assert().Equal(10, len(hosts))
		for _, host := range hosts {
			s.Assert().False(host.Provision)
		}

		err = s.db.ProvisionHosts(ns, true)
		s.Assert().NoError(err)

		hosts, err = s.db.FindHosts(ns)
		s.Assert().NoError(err)
		s.Assert().Equal(10, len(hosts))
		for _, host := range hosts {
			s.Assert().True(host.Provision)
		}
	}
}

func (s *DataStoreTestSuite) TestSetBootImage() {
	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := s.db.StoreHost(host)
		s.Assert().NoError(err)
	}

	ns, err := nodeset.NewNodeSet("tux-[05-14]")
	if s.Assert().NoError(err) {
		hosts, err := s.db.FindHosts(ns)
		s.Assert().NoError(err)
		s.Assert().Equal(10, len(hosts))
		for _, host := range hosts {
			s.Assert().Equal("", host.BootImage)
		}

		err = s.db.SetBootImage(ns, "centos7")
		s.Assert().NoError(err)

		hosts, err = s.db.FindHosts(ns)
		s.Assert().NoError(err)
		s.Assert().Equal(10, len(hosts))
		for _, host := range hosts {
			s.Assert().Equal("centos7", host.BootImage)
		}
	}
}

func (s *DataStoreTestSuite) TestBootImage() {
	image := tests.BootImageFactory.MustCreate().(*model.BootImage)

	image.ProvisionTemplates = map[string]string{
		"kickstart":    "kickstart.tmpl",
		"post-install": "post-install.tmpl",
	}

	err := s.db.StoreBootImage(image)
	s.Assert().NoError(err)

	testImage, err := s.db.LoadBootImage(image.Name)
	if s.Assert().NoError(err) {
		s.Assert().Equal(image.Name, testImage.Name)
		s.Assert().Contains(testImage.ProvisionTemplates, "post-install")
		s.Assert().Contains(testImage.ProvisionTemplates, "kickstart")
	}

	badimage := &model.BootImage{}
	err = s.db.StoreBootImage(badimage)
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, model.ErrInvalidData))
	}

	_, err = s.db.LoadBootImage("notfound")
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, model.ErrNotFound))
	}

	for i := 0; i < 5; i++ {
		image := tests.BootImageFactory.MustCreate().(*model.BootImage)
		err := s.db.StoreBootImage(image)
		s.Assert().NoError(err)
	}

	images, err := s.db.BootImages()
	if s.Assert().NoError(err) {
		s.Assert().Equal(6, len(images))
	}
}

func (s *DataStoreTestSuite) TestBootImageDelete() {
	image := tests.BootImageFactory.MustCreate().(*model.BootImage)

	err := s.db.StoreBootImage(image)
	s.Assert().NoError(err)

	testImage, err := s.db.LoadBootImage(image.Name)
	if s.Assert().NoError(err) {
		s.Assert().Equal(image.Name, testImage.Name)
	}

	err = s.db.DeleteBootImages([]string{testImage.Name})
	if s.Assert().NoError(err) {
		_, err = s.db.LoadBootImage(testImage.Name)
		if s.Assert().Error(err) {
			s.Assert().True(errors.Is(err, model.ErrNotFound))
		}
	}
}

func (s *DataStoreTestSuite) TestHostUpdate() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromID(host.ID.String())
	if s.Assert().NoError(err) {
		s.Assert().Equal(2, len(testHost.Interfaces))
	}

	// Store host with same name is update
	hostDup := tests.HostFactory.MustCreate().(*model.Host)
	hostDup.ID = host.ID
	hostDup.Name = host.Name
	err = s.db.StoreHost(hostDup)
	if s.Assert().NoError(err) {
		hosts, err := s.db.Hosts()
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(hosts))
	}

	// Store host with different name gets new ID
	hostDup = tests.HostFactory.MustCreate().(*model.Host)
	hostDup.ID = host.ID
	hostDup.Name = "cpn-new"
	err = s.db.StoreHost(hostDup)
	if s.Assert().NoError(err) {
		hosts, err := s.db.Hosts()
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(hosts))
		idCheck := ""
		for _, h := range hosts {
			s.Assert().NotEqual(idCheck, h.ID.String())
			idCheck = h.ID.String()
		}
	}
}

func (s *DataStoreTestSuite) TestHostDelete() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromID(host.ID.String())
	if s.Assert().NoError(err) {
		s.Assert().Equal(2, len(testHost.Interfaces))
	}

	ns, err := nodeset.NewNodeSet(testHost.Name)
	if s.Assert().NoError(err) {
		err := s.db.DeleteHosts(ns)
		s.Assert().NoError(err)

		_, err = s.db.LoadHostFromID(host.ID.String())
		if s.Assert().Error(err) {
			s.Assert().True(errors.Is(err, model.ErrNotFound))
		}
	}
}
