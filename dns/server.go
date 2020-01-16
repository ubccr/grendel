package dns

import (
	"github.com/miekg/dns"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/logger"
)

var log = logger.GetLogger("DNS")

type Server struct {
	Address string

	srv *dns.Server
}

func NewServer(client *client.Client, address string, ttl int) (*Server, error) {
	s := &Server{Address: address}

	s.srv = &dns.Server{Addr: address, Net: "udp"}
	h, err := NewHandler(client, uint32(ttl))
	if err != nil {
		return nil, err
	}

	s.srv.Handler = h

	return s, nil
}

func (s *Server) Serve() error {
	log.Infof("DNS server listening on: %s", s.Address)
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown() {
	s.srv.Shutdown()
}
