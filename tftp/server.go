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

package tftp

import (
	"context"
	"time"

	"github.com/pin/tftp"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
)

var log = logger.GetLogger("TFTP")

type Server struct {
	Address string
	DB      model.DataStore
	srv     *tftp.Server
}

func NewServer(db model.DataStore, address string) (*Server, error) {
	s := &Server{DB: db, Address: address}

	s.srv = tftp.NewServer(s.ReadHandler, nil)
	s.srv.SetTimeout(2 * time.Second)

	return s, nil
}

func (s *Server) Serve() error {
	log.Infof("Server listening on: %s", s.Address)
	return s.srv.ListenAndServe(s.Address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.srv.Shutdown()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}
}
