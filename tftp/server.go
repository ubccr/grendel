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
)

var log = logger.GetLogger("TFTP")

type Server struct {
	Address string
	srv     *tftp.Server
}

func NewServer(address string) (*Server, error) {
	s := &Server{Address: address}

	s.srv = tftp.NewServer(s.ReadHandler, nil)
	s.srv.SetTimeout(2 * time.Second)

	return s, nil
}

func (s *Server) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		log.Info("Shutting down TFTP server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.Shutdown(ctxShutdown); err != nil {
			log.Errorf("Failed shutting down TFTP server: %v", err)
		}
	}()

	log.Infof("Server listening on: %s", s.Address)
	if err := s.srv.ListenAndServe(s.Address); err != nil {
		return err
	}

	return nil

}

func (s *Server) shutdown() chan error {
	err := make(chan error, 1)
	defer close(err)
	err <- nil
	s.srv.Shutdown()

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	var ctxErr error
	select {
	case err := <-s.shutdown():
		ctxErr = err
	case <-ctx.Done():
		ctxErr = ctx.Err()
	}

	return ctxErr
}
