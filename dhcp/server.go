package dhcp

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/insomniacslk/dhcp/interfaces"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
)

var log = logger.GetLogger("DHCP")

type Server struct {
	ListenAddress net.IP
	ServerAddress net.IP
	IfIndex       int
	Hostname      string
	HTTPScheme    string
	Port          int
	HTTPPort      int
	PXEPort       int
	MTU           int
	ProxyOnly     bool
	ServePXE      bool
	DB            model.Datastore
	DNSServers    []net.IP
	LeaseTime     time.Duration
	srv           *server4.Server
	srvPXE        *server4.Server
}

func NewServer(db model.Datastore, address string, proxyOnly bool) (*Server, error) {
	s := &Server{DB: db, ProxyOnly: proxyOnly, ServePXE: true}

	if proxyOnly {
		log.Debugf("Running in ProxyOnly mode")
	}

	if address == "" {
		address = fmt.Sprintf("%s:%d", net.IPv4zero.String(), dhcpv4.ServerPort)
	}

	ipStr, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	s.Port = port

	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("Invalid IPv4 address: %s", ipStr)
	}

	if !ip.To4().Equal(net.IPv4zero) {
		s.ListenAddress = ip
		s.ServerAddress = ip
		return s, nil
	}

	// Attempt to discover server ip from interfaces

	intfs, err := interfaces.GetNonLoopbackInterfaces()
	if err != nil {
		return nil, err
	}

	serverIps := make([]net.IP, 0)

	for _, intf := range intfs {
		addrs, err := intf.Addrs()
		if err != nil {
			return nil, err
		}

		ips, err := dhcpv4.GetExternalIPv4Addrs(addrs)
		if err != nil {
			return nil, err
		}

		if len(ips) == 0 {
			continue
		}

		log.Debugf("Found IP(s) for interface %s: %v", intf.Name, ips)
		serverIps = append(serverIps, ips...)
		s.IfIndex = intf.Index
	}

	if len(serverIps) == 0 {
		return nil, errors.New("Failed to find server ip address from configured interfaces")
	}
	if len(serverIps) != 1 {
		//TODO add support for multiple interfaces
		return nil, fmt.Errorf("Multiple interfaces not supported yet: %#v", serverIps)
	}

	s.ServerAddress = serverIps[0]
	s.ListenAddress = ip

	return s, nil
}

func (s *Server) mainHandler4(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
	log.Debugf("Received DHCPv4 packet")
	log.Debugf(req.Summary())

	if req.OpCode != dhcpv4.OpcodeBootRequest {
		log.Warningf("not a BootRequest, ignoring")
		return
	}

	resp, err := dhcpv4.NewReplyFromRequest(req,
		dhcpv4.WithBroadcast(true),
		dhcpv4.WithServerIP(s.ServerAddress),
		dhcpv4.WithMessageType(dhcpv4.MessageTypeOffer),
		dhcpv4.WithOption(dhcpv4.OptClassIdentifier("PXEClient")),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(s.ServerAddress)),
	)
	if err != nil {
		log.Printf("DHCP failed to build reply: %v", err)
		return
	}

	if req.Options.Has(dhcpv4.OptionClientMachineIdentifier) {
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionClientMachineIdentifier, req.Options.Get(dhcpv4.OptionClientMachineIdentifier)))
	}

	switch mt := req.MessageType(); mt {
	case dhcpv4.MessageTypeDiscover:
		err := s.bootingHandler4(req, resp)
		if err != nil && s.ProxyOnly {
			log.Errorf("Failed to add boot options to DHCP request: %s", err)
			return
		}

		if !s.ProxyOnly {
			err := s.staticHandler4(req, resp)
			if err != nil {
				log.Errorf("Failed to find static DHCP lease: %s", err)
				return
			}
		}
	case dhcpv4.MessageTypeRequest:
		if s.ProxyOnly {
			return
		}

		err := s.staticAckHandler4(req, resp)
		if err != nil {
			log.Errorf("Failed to ack DHCP REQUEST: %s", err)
			return
		}
	default:
		log.Warnf("DHCP Unhandled message type: %v", mt)
		return
	}

	peer = &net.UDPAddr{IP: net.IPv4bcast, Port: dhcpv4.ClientPort}

	log.Debugf("Sending DHCPv4 packet response")
	log.Debugf(resp.Summary())

	if _, err := conn.WriteTo(resp.ToBytes(), peer); err != nil {
		log.Printf("DHCP conn.Write to %v failed: %v", peer, err)
	}
}

func (s *Server) Serve() error {
	if s.HTTPPort == 0 {
		s.HTTPPort = 80
	}
	if s.HTTPScheme == "" {
		s.HTTPScheme = "http"
	}
	if s.PXEPort == 0 {
		s.PXEPort = 4011
	}

	listener := &net.UDPAddr{
		IP:   s.ListenAddress,
		Port: s.Port,
	}
	srv, err := server4.NewServer("", listener, s.mainHandler4)
	if err != nil {
		return err
	}

	s.srv = srv

	log.Debugf("Server Address: %s", s.ServerAddress.String())

	if !s.ServePXE {
		return s.srv.Serve()
	}

	pxeListener := &net.UDPAddr{
		IP:   s.ListenAddress,
		Port: s.PXEPort,
	}

	srvPXE, err := server4.NewServer("", pxeListener, s.pxeHandler4)
	if err != nil {
		return err
	}

	s.srvPXE = srvPXE

	errs := make(chan error, 2)

	go func() { errs <- s.srv.Serve() }()
	go func() { errs <- s.srvPXE.Serve() }()

	err = <-errs
	s.Shutdown()

	return err
}

func (s *Server) Shutdown() {
	err := s.srv.Close()
	if err != nil {
		log.Errorf("Failed to close dhcp server: %s", err)
	}
	if !s.ServePXE {
		return
	}

	err = s.srvPXE.Close()
	if err != nil {
		log.Errorf("Failed to close pxe server: %s", err)
	}
}
