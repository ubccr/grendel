// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package storetest

import (
	"errors"
	"fmt"
	"math/rand"
	"net/netip"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

type StoreTestSuite struct {
	suite.Suite
	db store.Store
}

func (s *StoreTestSuite) SetStore(db store.Store) {
	s.db = db
}

func (s *StoreTestSuite) TestUser() {
	adminUsername := "admin"
	adminPassword := "SuperSecureAdminPassword1234!@#$"
	userUsername := "user"
	userPassword := "1234"

	defer s.db.Close()

	role, err := s.db.StoreUser(adminUsername, adminPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(model.RoleAdmin.String(), role)
	role, err = s.db.StoreUser(userUsername, userPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(model.RoleUser.String(), role)

	authenticated, role, err := s.db.VerifyUser("admin", adminPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(model.RoleAdmin.String(), role)
	s.Assert().Equal(true, authenticated)
	authenticated, role, err = s.db.VerifyUser("user", userPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(model.RoleUser.String(), role)
	s.Assert().Equal(true, authenticated)

	users, err := s.db.GetUsers()
	s.Assert().NoError(err)
	s.Assert().Equal(users[0].Username, "admin")
	s.Assert().Contains(users[1].Username, "user")

	err = s.db.UpdateUserRole("user", model.RoleReadOnly.String())
	s.Assert().NoError(err)
	authenticated, role, err = s.db.VerifyUser("user", userPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(role, model.RoleReadOnly.String())
	s.Assert().Equal(authenticated, true)

	err = s.db.DeleteUser("user")
	s.Assert().NoError(err)
	users, err = s.db.GetUsers()
	s.Assert().NoError(err)
	s.Assert().Equal(len(users), 1)
}

func (s *StoreTestSuite) TestHost() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromID(host.UID.String())
	if s.Assert().NoError(err) {
		s.Assert().Equal(host.Name, testHost.Name)
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
		s.Assert().True(errors.Is(err, store.ErrInvalidData))
	}

	_, err = s.db.LoadHostFromID("notfound")
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, store.ErrNotFound))
	}

	_, err = s.db.LoadHostFromName("notfound")
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, store.ErrNotFound))
	}

	_, err = s.db.LoadHostFromMAC("notfound")
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, store.ErrNotFound))
	}
}

func (s *StoreTestSuite) TestResolveIPv4() {
	testNames := []string{"test1.example.com", "cname.example.com"}

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.Interfaces[0].FQDN = strings.Join(testNames, ",")

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	for _, nm := range testNames {
		testIPs, err := s.db.ResolveIPv4(nm)
		if s.Assert().NoError(err) {
			if s.Assert().Equal(1, len(testIPs)) {
				s.Assert().Equal(host.Interfaces[0].AddrString(), testIPs[0].String())
			}
		}
	}

	names, err := s.db.ReverseResolve(host.Interfaces[0].AddrString())
	if s.Assert().NoError(err) {
		if s.Assert().Equal(1, len(names)) {
			s.Assert().Equal(testNames[0], names[0])
		}
	}
}

func (s *StoreTestSuite) TestResolveIPv4ExactMatch() {
	host := tests.HostFactory.MustCreate().(*model.Host)
	host.Interfaces[0].FQDN = "test1.example.com"
	host.Interfaces[1].FQDN = "xxx-test1.example.com"

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testIPs, err := s.db.ResolveIPv4("test1.example.com")
	if s.Assert().NoError(err) {
		if s.Assert().Equal(1, len(testIPs)) {
			s.Assert().Equal(host.Interfaces[0].AddrString(), testIPs[0].String())
		}
	}

	names, err := s.db.ReverseResolve(host.Interfaces[0].AddrString())
	if s.Assert().NoError(err) {
		if s.Assert().Equal(1, len(names)) {
			s.Assert().Equal("test1.example.com", names[0])
		}
	}
}

func (s *StoreTestSuite) TestReverseResolveIPv4() {
	hostA := tests.HostFactory.MustCreate().(*model.Host)
	hostA.Interfaces[0].IP = netip.MustParsePrefix("10.1.1.17/24")

	hostB := tests.HostFactory.MustCreate().(*model.Host)
	hostB.Interfaces[0].IP = netip.MustParsePrefix("10.1.1.174/24")

	err := s.db.StoreHost(hostA)
	s.Assert().NoError(err)

	err = s.db.StoreHost(hostB)
	s.Assert().NoError(err)

	testNames, err := s.db.ReverseResolve(hostA.Interfaces[0].AddrString())
	if s.Assert().NoError(err) {
		if s.Assert().Equal(1, len(testNames)) {
			s.Assert().Equal(hostA.Interfaces[0].FQDN, testNames[0])
		}
	}
}

func (s *StoreTestSuite) TestIfname() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromName(host.Name)
	if s.Assert().NoError(err) {
		s.Assert().Equal(host.Interfaces[0].Name, testHost.Interfaces[0].Name)
	}
}

func (s *StoreTestSuite) TestBonds() {
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

func (s *StoreTestSuite) TestHostList() {
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

func (s *StoreTestSuite) TestHostFind() {
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

func (s *StoreTestSuite) TestFindTags() {
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
		s.Assert().True(errors.Is(err, store.ErrNotFound))
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
		s.Assert().True(errors.Is(err, store.ErrNotFound))
	}

	ns, err = nodeset.NewNodeSet("tux-[05-06]")
	if s.Assert().NoError(err) {
		err := s.db.TagHosts(ns, []string{"pdu"})
		s.Assert().NoError(err)
	}

	ns, err = nodeset.NewNodeSet("tux-[07-08]")
	if s.Assert().NoError(err) {
		err := s.db.TagHosts(ns, []string{"switch"})
		s.Assert().NoError(err)
	}

	ns, err = s.db.MatchTags([]string{"harkness", "pdu"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(2, ns.Len())
	}

	ns, err = s.db.FindTags([]string{"harkness", "pdu"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(4, ns.Len())
	}
}

func (s *StoreTestSuite) TestUpdateTags() {
	size := 5
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		host.Tags = []string{"rack10", "hpc"}
		err := s.db.StoreHost(host)
		s.Assert().NoError(err)
	}

	ns, err := s.db.FindTags([]string{"rack10"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(size, ns.Len())
	}

	testHost, err := s.db.LoadHostFromName("tux-01")
	if s.Assert().NoError(err) {
		s.Assert().Equal(2, len(testHost.Tags))
	}

	// Update Tags
	testHost.Tags = []string{"rack10", "faculty"}
	err = s.db.StoreHost(testHost)
	s.Assert().NoError(err)

	ns, err = s.db.FindTags([]string{"rack10"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(size, ns.Len())
	}

	ns, err = s.db.FindTags([]string{"faculty"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(1, ns.Len())
	}

	ns, err = s.db.FindTags([]string{"hpc"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(size-1, ns.Len())
	}

	// Delete Tags
	testHost.Tags = []string{"test"}
	err = s.db.StoreHost(testHost)
	s.Assert().NoError(err)

	ns, err = s.db.FindTags([]string{"rack10"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(size-1, ns.Len())
	}

	_, err = s.db.FindTags([]string{"faculty"})
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, store.ErrNotFound))
	}

	ns, err = s.db.FindTags([]string{"hpc"})
	if s.Assert().NoError(err) {
		s.Assert().Equal(size-1, ns.Len())
	}
}

func (s *StoreTestSuite) TestProvision() {
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

func (s *StoreTestSuite) TestSetBootImage() {
	image := tests.BootImageFactory.MustCreate().(*model.BootImage)
	image.Name = "centos7"
	err := s.db.StoreBootImage(image)
	s.Assert().NoError(err)

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

func (s *StoreTestSuite) TestBootImage() {
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
		s.Assert().True(errors.Is(err, store.ErrInvalidData))
	}

	_, err = s.db.LoadBootImage("notfound")
	if s.Assert().Error(err) {
		s.Assert().True(errors.Is(err, store.ErrNotFound))
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

func (s *StoreTestSuite) TestBootImageDelete() {
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
			s.Assert().True(errors.Is(err, store.ErrNotFound))
		}
	}
}

func (s *StoreTestSuite) TestBootImageUpdate() {
	image := tests.BootImageFactory.MustCreate().(*model.BootImage)

	image.ProvisionTemplates = map[string]string{
		"kickstart":    "kickstart.tmpl",
		"post-install": "post-install.tmpl",
	}

	err := s.db.StoreBootImage(image)
	s.Assert().NoError(err)

	// try storing host with no changes
	err = s.db.StoreBootImage(image)
	s.Assert().NoError(err)

	testImage, err := s.db.LoadBootImage(image.Name)
	if s.Assert().NoError(err) {
		s.Assert().Equal(image.Name, testImage.Name)
		s.Assert().Equal(1, len(testImage.InitrdPaths))
		s.Assert().Contains(testImage.ProvisionTemplates, "post-install")
		s.Assert().Contains(testImage.ProvisionTemplates, "kickstart")
	}

	testImage.ProvisionTemplates = map[string]string{
		"kickstart": "kickstart.tmpl",
		"test":      "test.tmpl",
	}
	testImage.InitrdPaths = []string{
		"/path/1",
		"/path/2",
	}
	testImage.Verify = true

	err = s.db.StoreBootImage(testImage)
	s.Assert().NoError(err)

	testImage, err = s.db.LoadBootImage(testImage.Name)
	if s.Assert().NoError(err) {
		s.Assert().Equal(image.Name, testImage.Name)
		s.Assert().Equal(2, len(testImage.InitrdPaths))
		s.Assert().NotContains(testImage.ProvisionTemplates, "post-install")
		s.Assert().Contains(testImage.ProvisionTemplates, "kickstart")
		s.Assert().Equal(true, testImage.Verify)
	}
}

func (s *StoreTestSuite) TestHostUpdate() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromID(host.UID.String())
	if s.Assert().NoError(err) {
		s.Assert().Equal(2, len(testHost.Interfaces))
	}

	// Store host with same id is an update
	hostDup := tests.HostFactory.MustCreate().(*model.Host)
	hostDup.ID = host.ID
	hostDup.Name = "new name"
	err = s.db.StoreHost(hostDup)
	if s.Assert().NoError(err) {
		hosts, err := s.db.Hosts()
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(hosts))
	}

	// Store host with same name different id is an err
	hostDup = tests.HostFactory.MustCreate().(*model.Host)
	hostDup.UID = host.UID
	hostDup.Name = host.Name
	err = s.db.StoreHost(hostDup)
	if s.Assert().Error(err) {
		hosts, err := s.db.Hosts()
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(hosts))
	}

	// Store host with different name gets new ID
	hostDup = tests.HostFactory.MustCreate().(*model.Host)
	hostDup.UID.Set("")
	hostDup.Name = "cpn-new"
	err = s.db.StoreHost(hostDup)
	if s.Assert().NoError(err) {
		hosts, err := s.db.Hosts()
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(hosts))
		idCheck := ""
		for _, h := range hosts {
			s.Assert().NotEqual(idCheck, h.UID.String())
			idCheck = h.UID.String()
		}
	}
}

func (s *StoreTestSuite) TestHostNics() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromID(host.UID.String())
	if s.Assert().NoError(err) {
		s.Assert().Equal(2, len(testHost.Interfaces))
	}

	// Update nic
	testHost.Interfaces[0].Name = "testifname0"
	err = s.db.StoreHost(testHost)
	if s.Assert().NoError(err) {
		h, err := s.db.LoadHostFromID(testHost.UID.String())
		if s.Assert().NoError(err) {
			s.Assert().Equal(testHost.Interfaces[0].Name, h.Interfaces[0].Name)
			s.Assert().Equal(len(testHost.Interfaces), len(h.Interfaces))
		}
	}

	// Delete nic
	testHost.Interfaces = testHost.Interfaces[1:len(testHost.Interfaces)]
	err = s.db.StoreHost(testHost)
	if s.Assert().NoError(err) {
		h, err := s.db.LoadHostFromID(testHost.UID.String())
		if s.Assert().NoError(err) {
			s.Assert().Equal(len(testHost.Interfaces), len(h.Interfaces))
		}
	}

	// Add nic
	newNic := tests.NetInterfaceFactory.MustCreate().(*model.NetInterface)
	testHost.Interfaces = append(testHost.Interfaces, newNic)
	err = s.db.StoreHost(testHost)
	if s.Assert().NoError(err) {
		h, err := s.db.LoadHostFromID(testHost.UID.String())
		if s.Assert().NoError(err) {
			s.Assert().Equal(len(testHost.Interfaces), len(h.Interfaces))
		}
	}

}

func (s *StoreTestSuite) TestHostDelete() {
	host := tests.HostFactory.MustCreate().(*model.Host)

	err := s.db.StoreHost(host)
	s.Assert().NoError(err)

	testHost, err := s.db.LoadHostFromID(host.UID.String())
	if s.Assert().NoError(err) {
		s.Assert().Equal(2, len(testHost.Interfaces))
	}

	ns, err := nodeset.NewNodeSet(testHost.Name)
	if s.Assert().NoError(err) {
		err := s.db.DeleteHosts(ns)
		s.Assert().NoError(err)

		_, err = s.db.LoadHostFromID(host.UID.String())
		if s.Assert().Error(err) {
			s.Assert().True(errors.Is(err, store.ErrNotFound))
		}
	}
}

func (s *StoreTestSuite) TestRestore() {
	size := 10
	adminUsername := "admin"
	adminPassword := "pass1234123"

	dump := model.DataDump{
		Users:  make([]model.User, 0),
		Hosts:  make(model.HostList, 0),
		Images: make(model.BootImageList, 0),
	}

	role, err := s.db.StoreUser(adminUsername, adminPassword)
	s.Assert().NoError(err)
	s.Assert().Equal("admin", role)
	role, err = s.db.StoreUser("user", "test1234")
	s.Assert().NoError(err)
	s.Assert().Equal("user", role)

	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		dump.Hosts = append(dump.Hosts, host)
	}

	for i := 0; i < size; i++ {
		image := tests.BootImageFactory.MustCreate().(*model.BootImage)
		dump.Images = append(dump.Images, image)
	}

	users, err := s.db.GetUsers()
	s.Assert().NoError(err)
	for _, u := range users {
		dump.Users = append(dump.Users, u)
	}

	err = s.db.RestoreFrom(dump)
	s.Assert().NoError(err)

	hostList, err := s.db.Hosts()
	s.Assert().NoError(err)
	s.Assert().Equal(len(dump.Hosts), len(hostList))

	imageList, err := s.db.BootImages()
	s.Assert().NoError(err)
	s.Assert().Equal(len(dump.Images), len(imageList))

	users2, err := s.db.GetUsers()
	s.Assert().NoError(err)
	s.Assert().Equal(len(dump.Users), len(users2))

	authenticated, role, err := s.db.VerifyUser(adminUsername, adminPassword)
	s.Assert().NoError(err)
	s.Assert().Equal("admin", role)
	s.Assert().Equal(true, authenticated)
}

func (s *StoreTestSuite) BenchmarkWriteNodes(size int, b *testing.B) {
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := s.db.StoreHosts(hosts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func (s *StoreTestSuite) BenchmarkWriteSingleNode(size int, b *testing.B) {
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < size; i++ {
				err := s.db.StoreHost(hosts[i])
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func (s *StoreTestSuite) BenchmarkReadAll(size int, b *testing.B) {
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	err := s.db.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			list, err := s.db.Hosts()
			if err != nil {
				b.Fatal(err)
			}
			if len(list) != size {
				b.Fatalf("wrong size: expected %d got %d", size, len(list))
			}
		}
	})
}

func (s *StoreTestSuite) BenchmarkFind(size int, b *testing.B) {
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%04d", i)
		hosts[i] = host
	}

	err := s.db.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := rand.Intn(int(size / 2))
			start := rand.Intn(int(size / 2))
			end := start + n
			if end > size-1 {
				end = size - 1
			}

			n = end - start

			ns, err := nodeset.NewNodeSet(fmt.Sprintf("tux-[%04d-%04d]", start, end))
			if err != nil {
				b.Fatal(err)
			}

			hosts, err := s.db.FindHosts(ns)
			if err != nil {
				b.Fatal(err)
			}

			if len(hosts) != n+1 {
				b.Fatalf("wrong host count found: expected %d got %d", n+1, len(hosts))
			}
		}
	})
}

func (s *StoreTestSuite) BenchmarkRandomWrites(size int, b *testing.B) {
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := rand.Intn(int(size / 2))
			start := rand.Intn(int(size / 2))
			end := start + n
			if end > size-1 {
				end = size - 1
			}

			picks := hosts[start:end]
			hosts := make(model.HostList, len(picks))
			for i, h := range picks {
				hosts[i] = h
			}
			err := s.db.StoreHosts(hosts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func (s *StoreTestSuite) BenchmarkRandomReads(size int, b *testing.B) {
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	err := s.db.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pick := hosts[rand.Intn(size)]
			_, err := s.db.LoadHostFromID(pick.UID.String())
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func (s *StoreTestSuite) BenchmarkResolveIP(size int, b *testing.B) {
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	err := s.db.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pick := hosts[rand.Intn(size)]
			ips, err := s.db.ResolveIPv4(pick.Interfaces[0].FQDN)
			if err != nil {
				b.Fatal(err)
			}
			if len(ips) != 1 {
				b.Fatalf("IPs not found")
			}
		}
	})
}

func (s *StoreTestSuite) BenchmarkReverseResolve(size int, b *testing.B) {
	hosts := make(model.HostList, size)
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		hosts[i] = host
	}

	err := s.db.StoreHosts(hosts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pick := hosts[rand.Intn(size)]
			names, err := s.db.ReverseResolve(pick.Interfaces[0].AddrString())
			if err != nil {
				b.Fatal(err)
			}
			if len(names) != len(strings.Split(pick.Interfaces[0].FQDN, ",")) {
				b.Fatalf("wrong fqdn expected %s got %#v", pick.Interfaces[0].FQDN, names)
			}
		}
	})
}
