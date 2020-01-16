package dhcp

import (
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/tors"
)

var (
	nodeNumberRegexp = regexp.MustCompile(`(\d+)$`)
)

type discovery struct {
	nodeset  *nodeset.NodeSetIterator
	seen     map[string]bool
	count    int
	macTable tors.MACTable
	subnet   net.IP
	netmask  net.IPMask
}

func RunDiscovery(address, nodestr string, subnet net.IP, netmask net.IPMask, switchClient tors.NetworkSwitch) error {
	if address == "" {
		address = fmt.Sprintf("%s:%d", net.IPv4zero.String(), dhcpv4.ServerPort)
	}

	var macTable tors.MACTable
	if switchClient != nil {
		mt, err := switchClient.GetMACTable()
		if err != nil {
			return err
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

	nodeset, err := nodeset.NewNodeSet(nodestr)
	if err != nil {
		return err
	}

	d := &discovery{
		nodeset:  nodeset.Iterator(),
		seen:     make(map[string]bool),
		macTable: macTable,
		subnet:   subnet,
		netmask:  netmask,
	}

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

	var entry *tors.MACTableEntry

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

	ip := d.subnet.Mask(d.netmask)
	if entry != nil {
		ip[3] += uint8(entry.Port)
	} else {
		matches := nodeNumberRegexp.FindStringSubmatch(d.nodeset.Value())
		if len(matches) != 2 {
			log.Errorf("node doesn't end in number. failed to generate IP address: %s", d.nodeset.Value())
			return
		}
		num, _ := strconv.Atoi(matches[1])
		ip[3] += uint8(num)
	}

	fmt.Printf("%s\t%s\t%s\n", d.nodeset.Value(), req.ClientHWAddr, ip.String())
}
