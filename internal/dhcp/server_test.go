// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build integration

package dhcp

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"testing"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/nclient4"
	"github.com/insomniacslk/dhcp/iana"
	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/model"
)

const (
	address = "0.0.0.0"
	port    = 67
	dbpath  = ":memory:"
)

var (
	clientIface  = "lo"
	clientPort   = 68
	clientHwAddr = net.HardwareAddr{1, 2, 3, 4, 5, 6}
	clientIpAddr = netip.MustParsePrefix("10.1.0.1/24")

	serverAddr = ""
)

func newTestDB(t *testing.T) store.Store {

	db, err := sqlstore.New(dbpath)
	if err != nil {
		t.Fatal(err)
	}

	host := model.Host{
		Name:      "test",
		Provision: true,
		Interfaces: []*model.NetInterface{
			{
				MAC: clientHwAddr,
				IP:  clientIpAddr,
			},
		},
		BootImage: "test-img",
	}
	img := model.BootImage{
		Name:        "test-img",
		KernelPath:  "/test",
		InitrdPaths: []string{"/test"},
	}

	err = db.StoreHost(&host)
	if err != nil {
		t.Fatal(err)
	}
	err = db.StoreBootImage(&img)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

// TestDhcpServer needs the cap_net_raw set on the test binary
func TestDhcpServer(t *testing.T) {
	assert := assert.New(t)
	db := newTestDB(t)

	s, err := NewServer(db, fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		t.Fatal(err)
	}

	eIp, err := util.GetFirstExternalIPFromInterfaces()
	serverAddr = eIp.String()

	go func() {
		err = s.Serve()
		assert.NoError(err)
	}()

	defer func() {
		err = s.Shutdown(context.Background())
		assert.NoError(err)
	}()

	conn, err := nclient4.NewRawUDPConn(clientIface, clientPort)
	if err != nil {
		t.Fatal(err)
	}

	c, err := nclient4.NewWithConn(conn, clientHwAddr)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	offer, err := c.Request(ctx, dhcpv4.WithOption(dhcpv4.OptClientArch(iana.EFI_X86_64)))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(offer.ACK.ServerIPAddr.String(), serverAddr)
	assert.Equal(offer.ACK.OpCode, dhcpv4.OpcodeBootReply)
	assert.Equal(offer.ACK.HWType, iana.HWTypeEthernet)
	assert.Equal(offer.ACK.ClientIPAddr.String(), "0.0.0.0")
	assert.Equal(offer.ACK.YourIPAddr.String(), clientIpAddr.Addr().String())
	assert.Equal(offer.ACK.ServerIPAddr.String(), serverAddr)
	assert.Equal(offer.ACK.GatewayIPAddr.String(), "0.0.0.0")
	assert.Equal(offer.ACK.ClientHWAddr, clientHwAddr)
	assert.Equal(offer.ACK.BootFileName, "")
	assert.Equal(offer.ACK.Options.Get(dhcpv4.OptionDHCPMessageType), dhcpv4.MessageTypeAck.ToBytes())
	assert.Equal(offer.ACK.Options.Get(dhcpv4.OptionClassIdentifier), []byte("PXEClient"))

	assert.Equal(offer.Offer.ServerIPAddr.String(), serverAddr)
	assert.Equal(offer.Offer.OpCode, dhcpv4.OpcodeBootReply)
	assert.Equal(offer.Offer.HWType, iana.HWTypeEthernet)
	assert.Equal(offer.Offer.ClientIPAddr.String(), "0.0.0.0")
	assert.Equal(offer.Offer.YourIPAddr.String(), clientIpAddr.Addr().String())
	assert.Equal(offer.Offer.ServerIPAddr.String(), serverAddr)
	assert.Equal(offer.Offer.GatewayIPAddr.String(), "0.0.0.0")
	assert.Equal(offer.Offer.ClientHWAddr, clientHwAddr)
	assert.Equal(offer.Offer.BootFileName, "")
	assert.Equal(offer.Offer.Options.Get(dhcpv4.OptionDHCPMessageType), dhcpv4.MessageTypeOffer.ToBytes())
	assert.Equal(offer.Offer.Options.Get(dhcpv4.OptionServerIdentifier), []uint8(net.ParseIP(serverAddr).To4()))
	assert.Equal(offer.Offer.Options.Get(dhcpv4.OptionClassIdentifier), []byte("PXEClient"))
	assert.Equal(offer.Offer.Options.Get(dhcpv4.OptionTFTPServerName), []byte(serverAddr))
	assert.NotEqual(offer.Offer.Options.Get(dhcpv4.OptionBootfileName), []byte(""))

	// inform, err := c.Inform(ctx, net.ParseIP("10.1.0.1"))
	// if err != nil {
	// 	t.Fatal(err)
	// }

	err = c.Release(offer)
	if err != nil {
		t.Fatal(err)
	}

}
