// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/util"
	"golang.org/x/sys/unix"
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
	socketServer  *http.Server
	SwaggerUI     bool
	CORS          bool
}

func NewServer(db store.Store, socket, address string) (*Server, error) {
	s := &Server{Scheme: "http", DB: db, SocketPath: socket}

	if address == "" {
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
	s.server = fuego.NewServer(
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

	s.server.OpenAPI.Description().Info.Title = "Grendel API"
	s.server.OpenAPI.Description().Info.Description = "OpenAPI spec for the Grendel API"
	s.server.OpenAPI.Description().Info.Version = "0.2.0"

	h, err := NewHandler(s.DB)
	if err != nil {
		return err
	}

	h.SetupRoutes(s.server)

	// Fix >30s handlers from returning an empty body
	s.server.Server.WriteTimeout = time.Minute * 5

	// UNIX listener
	if s.SocketPath != "" {
		os.Remove(s.SocketPath)
		unixListener, err := net.Listen("unix", s.SocketPath)
		if err != nil {
			return err
		}

		if err := os.Chmod(s.SocketPath, 0770); err != nil {
			return err
		}
		log.Infof("Listening on %s://%s", "unix", s.SocketPath)
		s.socketServer = &http.Server{
			Handler:     blockWebUI(s.server.Mux),
			ConnContext: unixPeerConnContext,
		}
		go func() {
			err := s.socketServer.Serve(unixListener)
			if err != nil {
				log.Errorf("failed to start unix listener: %s", err)
			}
		}()
	}

	// TCP listener
	if s.ListenAddress != nil {
		s.server.Addr = fmt.Sprintf("%s:%d", s.ListenAddress, s.Port)

		if s.CertFile != "" && s.KeyFile != "" {
			s.Scheme = "https"
			log.Infof("Listening on %s://%s:%d", s.Scheme, s.ListenAddress, s.Port)
			return s.server.RunTLS(s.CertFile, s.KeyFile)
		} else {
			return s.server.Run()
		}

	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(context.TODO())

	}
	return errors.New("failed to create api server")
}

func blockWebUI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/ui") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Set username context on UNIX listener requests
func unixPeerConnContext(ctx context.Context, c net.Conn) context.Context {
	uc, ok := c.(*net.UnixConn)
	if !ok {
		return ctx
	}
	raw, err := uc.SyscallConn()
	if err != nil {
		return ctx
	}

	var cred *unix.Ucred
	var sockErr error
	err = raw.Control(func(fd uintptr) {
		cred, sockErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	if err != nil || sockErr != nil {
		return ctx
	}

	name := strconv.FormatUint(uint64(cred.Uid), 10)
	u, err := user.LookupId(name)
	if err == nil {
		name = u.Username
	}

	return context.WithValue(ctx, ContextKeyUsername, "unix:"+name)
}
