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
