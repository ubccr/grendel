// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/util"
)

var log = logger.GetLogger("API")

type Server struct {
	ListenAddress net.IP
	ServerAddress net.IP
	SocketPath    string
	Port          int
	Scheme        string
	KeyFile       string
	CertFile      string
	Hostname      string
	DB            store.Store
	server        *fuego.Server
	SwaggerUI     bool
	CORS          bool
}

func NewServer(db store.Store, socket, address string) (*Server, error) {
	s := &Server{Scheme: "http", DB: db, SocketPath: socket}

	if socket != "" {
		return s, nil
	}

	shost, sport, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if shost == "" {
		shost = net.IPv4zero.String()
	}

	port := DefaultPort
	if sport != "" {
		var err error
		port, err = strconv.Atoi(sport)
		if err != nil {
			return nil, err
		}
	}

	s.Port = port

	ip := net.ParseIP(shost)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("Invalid IPv4 address: %s", shost)
	}

	s.ListenAddress = ip

	if !ip.To4().Equal(net.IPv4zero) {
		s.ServerAddress = ip
		return s, nil
	}

	ipaddr, err := util.GetFirstExternalIPFromInterfaces()
	if err != nil {
		return nil, err
	}

	s.ServerAddress = ipaddr

	return s, nil
}

func (s *Server) Serve() error {

	var listen func(*fuego.Server)

	if s.SocketPath != "" {
		os.Remove(s.SocketPath)
		unixListener, err := net.Listen("unix", s.SocketPath)
		if err != nil {
			return err
		}

		if err := os.Chmod(s.SocketPath, 0770); err != nil {
			return err
		}
		log.Printf("Listening on unix domain socket: %s", s.SocketPath)
		listen = fuego.WithListener(unixListener)
	} else {
		addr := fmt.Sprintf("%s:%d", s.ListenAddress, s.Port)

		listen = fuego.WithAddr(addr)
	}

	s.server = fuego.NewServer(
		listen,
		fuego.WithEngineOptions(
			fuego.WithOpenAPIGeneratorOptions(
				openapi3gen.UseAllExportedFields(),
				openapi3gen.SchemaCustomizer(schemaCustomizer()),
			),
			fuego.WithOpenAPIConfig(setupOpenapiConfig(s.SwaggerUI)),
			fuego.WithErrorHandler(ErrorHandler),
		),
		fuego.WithErrorSerializer(ErrorSerializer),
		fuego.WithGlobalMiddlewares(
			corsMiddleware(s.CORS),
			logMiddleware,
		),
		fuego.WithSecurity(setupSecurity()),
	)

	h, err := NewHandler(s.DB)
	if err != nil {
		return err
	}

	h.SetupRoutes(s.server)
	if s.CertFile != "" && s.KeyFile != "" {
		s.Scheme = "https"
		log.Infof("Listening on %s://%s:%d", s.Scheme, s.ListenAddress, s.Port)
		return s.server.RunTLS(s.CertFile, s.KeyFile)
	}

	if s.SocketPath == "" {
		log.Infof("Listening on %s://%s:%d", s.Scheme, s.ListenAddress, s.Port)
	}
	return s.server.Run()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(context.TODO())

	}
	return errors.New("failed to create api server")
}
