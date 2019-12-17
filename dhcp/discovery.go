package dhcp

import (
	"fmt"
	"net"
	"strconv"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/tor"
)

type discovery struct {
	nodeset  *nodeset.NodeSet
	seen     map[string]bool
	count    int
	macTable tor.MACTable
}

func RunDiscovery(db model.Datastore, address, prefix, pattern string, switchClient tor.NetworkSwitch) error {
	if address == "" {
		address = fmt.Sprintf("%s:%d", net.IPv4zero.String(), dhcpv4.ServerPort)
	}

	var macTable tor.MACTable
	if switchClient != nil {
		mt, err := switchClient.GetMACTable()
		if err != nil {
			return nil
		}
		macTable = mt
	}

	ipStr, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("Invalid IPv4 address: %s", ipStr)
	}

	listener := &net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	nodeset, err := nodeset.NewNodeSet(prefix, pattern)
	if err != nil {
		return err
	}

	d := &discovery{nodeset: nodeset, seen: make(map[string]bool), macTable: macTable}

	log.Infof("Running auto discovery for nodeset size: %d", d.nodeset.Len())

	srv, err := server4.NewServer("", listener, d.discoveryHandler4)
	if err != nil {
		return err
	}

	return srv.Serve()
}

func (d *discovery) discoveryHandler4(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
	log.Debugf("Discovery received DHCPv4 packet")
	log.Debugf(req.Summary())

	if req.OpCode != dhcpv4.OpcodeBootRequest {
		log.Warningf("not a BootRequest, ignoring")
		return
	}

	if req.MessageType() != dhcpv4.MessageTypeDiscover {
		log.Warnf("Discovery unhandled message type: %v", req.MessageType())
		return
	}

	if _, ok := d.seen[req.ClientHWAddr.String()]; ok {
		log.Infof("Already seen mac address. skipping: %s", req.ClientHWAddr)
		return
	}

	var entry *tor.MACTableEntry

	if d.macTable != nil {
		if _, ok := d.macTable[req.ClientHWAddr.String()]; !ok {
			log.Infof("mac address does not exist in switch mac table. skipping: %s", req.ClientHWAddr)
			return
		}

		entry = d.macTable[req.ClientHWAddr.String()]
	}

	if !d.nodeset.Next() {
		log.Errorf("No more values in nodeset")
		return
	}

	d.seen[req.ClientHWAddr.String()] = true

	if entry != nil {
		fmt.Printf("%s\t%s\t%d\n", d.nodeset.Value(), req.ClientHWAddr, entry.Port)
	} else {
		fmt.Printf("%s\t%s\n", d.nodeset.Value(), req.ClientHWAddr)
	}
}