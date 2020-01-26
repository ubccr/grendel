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
