package tftp

import (
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

func (s *Server) Serve() error {
	log.Infof("TFTP server listening on: %s", s.Address)
	return s.srv.ListenAndServe(s.Address)
}

func (s *Server) Shutdown() {
	s.srv.Shutdown()
}
