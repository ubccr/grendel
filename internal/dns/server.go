// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package dns

import (
	"context"

	"github.com/miekg/dns"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/pkg/model"
)

var log = logger.GetLogger("DNS")

type Server struct {
	Address string

	srv *dns.Server
}

func NewServer(db model.DataStore, address string, ttl int) (*Server, error) {
	s := &Server{Address: address}

	s.srv = &dns.Server{Addr: address, Net: "udp"}
	h, err := NewHandler(db, uint32(ttl))
	if err != nil {
		return nil, err
	}

	s.srv.Handler = h

	return s, nil
}

func (s *Server) Serve() error {
	log.Infof("Server listening on: %s", s.Address)
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.ShutdownContext(ctx)
}
