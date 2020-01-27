package dhcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/util"
)

var log = logger.GetLogger("DHCP")

type Server struct {
	ListenAddress    net.IP
	ServerAddress    net.IP
	Hostname         string
	HTTPScheme       string
	Port             int
	HTTPPort         int
	MTU              int
	ProxyOnly        bool
	DB               model.Datastore
	DNSServers       []net.IP
	DomainSearchList []string
	LeaseTime        time.Duration
	srv              *server4.Server
}

func NewServer(db model.Datastore, address string) (*Server, error) {
	s := &Server{DB: db}

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

	s.ListenAddress = ip

	if !ip.To4().Equal(net.IPv4zero) {
		s.ServerAddress = ip
		return s, nil
	}

	ipaddr, err := util.GetInterfaceIP()
	if err != nil {
		return nil, err
	}

	s.ServerAddress = ipaddr

	return s, nil
}

func (s *Server) mainHandler4(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
	if req.OpCode != dhcpv4.OpcodeBootRequest {
		log.Debugf("Ignoring not a BootRequest")
		return
	}

	host, err := s.DB.LoadHostFromMAC(req.ClientHWAddr.String())
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			log.Debugf("Ignoring unknown client mac address: %s", req.ClientHWAddr)
		} else {
			log.Errorf("Failed to find host from database: %s", err)
		}
		return
	}

	resp, err := dhcpv4.NewReplyFromRequest(req,
		//dhcpv4.WithBroadcast(true),
		dhcpv4.WithServerIP(s.ServerAddress),
		dhcpv4.WithMessageType(dhcpv4.MessageTypeOffer),
		dhcpv4.WithOption(dhcpv4.OptClassIdentifier("PXEClient")),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(s.ServerAddress)),
	)
	if err != nil {
		log.Printf("DHCP failed to build reply: %v", err)
		return
	}

	// Copy hop count? is this needed?
	resp.HopCount = req.HopCount

	if req.Options.Has(dhcpv4.OptionClientMachineIdentifier) {
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionClientMachineIdentifier, req.Options.Get(dhcpv4.OptionClientMachineIdentifier)))
	}

	switch mt := req.MessageType(); mt {
	case dhcpv4.MessageTypeDiscover:
		err := s.bootingHandler4(host, req, resp)
		if err != nil && s.ProxyOnly {
			log.WithFields(logrus.Fields{
				"mac":     req.ClientHWAddr.String(),
				"host_id": host.ID.String(),
				"err":     err,
			}).Error("Failed to add boot options to DHCP request")
			return
		}

		if !s.ProxyOnly {
			err := s.staticHandler4(host, req, resp)
			if err != nil {
				return
			}
		}
	case dhcpv4.MessageTypeRequest:
		if s.ProxyOnly {
			return
		}

		err := s.staticAckHandler4(host, req, resp)
		if err != nil {
			log.Errorf("Failed to ack DHCP REQUEST: %s", err)
			return
		}
	default:
		log.Warnf("DHCP Unhandled message type: %v", mt)
		log.Debugf(resp.Summary())
		return
	}

	peer = &net.UDPAddr{IP: net.IPv4bcast, Port: dhcpv4.ClientPort}
	if !req.GatewayIPAddr.IsUnspecified() {
		peer = &net.UDPAddr{IP: req.GatewayIPAddr, Port: dhcpv4.ServerPort}
		resp.SetBroadcast()
	} else if req.ClientIPAddr != nil && !req.ClientIPAddr.Equal(net.IPv4zero) {
		peer = &net.UDPAddr{IP: req.ClientIPAddr, Port: dhcpv4.ClientPort}
		resp.SetUnicast()
	}

	log.Debugf("Sending DHCPv4 packet response")
	log.Debugf(resp.Summary())

	if _, err := conn.WriteTo(resp.ToBytes(), peer); err != nil {
		log.Printf("DHCP conn.Write to %v failed: %v", peer, err)
	}
}

func (s *Server) Serve(ctx context.Context) error {
	if s.HTTPPort == 0 {
		s.HTTPPort = 80
	}

	if s.HTTPScheme == "" {
		s.HTTPScheme = "http"
	}

	if s.ProxyOnly {
		log.Infof("Running in ProxyOnly mode")
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

	go func() {
		<-ctx.Done()

		log.Info("Shutting down DHCP server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.Shutdown(ctxShutdown); err != nil {
			log.Errorf("Failed shutting down DHCP server: %v", err)
		}
	}()

	log.Infof("Server listening on: %s:%d", s.ListenAddress, s.Port)

	if err := s.srv.Serve(); err != nil {
		log.Debugf("Error serving DHCP: %v", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	errs := make(chan error, 1)
	defer close(errs)

	go func() {
		errs <- s.srv.Close()
	}()

	var ctxErr error
	select {
	case err := <-errs:
		ctxErr = err
	case <-ctx.Done():
		ctxErr = ctx.Err()
	}

	return ctxErr
}
