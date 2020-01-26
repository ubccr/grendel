package dns

import (
	"context"
	"time"

	"github.com/miekg/dns"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
)

var log = logger.GetLogger("DNS")

type Server struct {
	Address string

	srv *dns.Server
}

func NewServer(db model.Datastore, address string, ttl int) (*Server, error) {
	s := &Server{Address: address}

	s.srv = &dns.Server{Addr: address, Net: "udp"}
	h, err := NewHandler(db, uint32(ttl))
	if err != nil {
		return nil, err
	}

	s.srv.Handler = h

	return s, nil
}

func (s *Server) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		log.Info("Shutting down DNS server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.Shutdown(ctxShutdown); err != nil {
			log.Errorf("Failed shutting down DNS server: %v", err)
		}
	}()

	log.Infof("Server listening on: %s", s.Address)
	if err := s.srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.ShutdownContext(ctx)
}
