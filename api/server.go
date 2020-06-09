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

package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/util"
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
	DB            model.DataStore
	httpServer    *http.Server
}

func newEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Logger = EchoLogger()
	e.Validator = &CustomValidator{validator: validator.New()}

	return e

}

func NewServer(db model.DataStore, socket, address string) (*Server, error) {
	s := &Server{DB: db, SocketPath: socket}

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

func HTTPErrorHandler(err error, c echo.Context) {
	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code == http.StatusNotFound {
			log.WithFields(logrus.Fields{
				"path": c.Request().URL,
				"ip":   c.RealIP(),
			}).Warn("Requested path not found")
		} else {
			log.WithFields(logrus.Fields{
				"code": he.Code,
				"err":  he.Internal,
				"path": c.Request().URL,
				"ip":   c.RealIP(),
			}).Error(he.Message)
		}
	} else {
		log.WithFields(logrus.Fields{
			"err":  err,
			"path": c.Request().URL,
			"ip":   c.RealIP(),
		}).Error("HTTP Error")
	}

	c.Echo().DefaultHTTPErrorHandler(err, c)
}

func (s *Server) Serve() error {
	e := newEcho()

	h, err := NewHandler(s.DB)
	if err != nil {
		return err
	}

	h.SetupRoutes(e)

	httpServer := &http.Server{
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	if s.SocketPath != "" {
		os.Remove(s.SocketPath)
		unixListener, err := net.Listen("unix", s.SocketPath)
		if err != nil {
			return err
		}
		e.Listener = unixListener
		s.Scheme = "http"
		log.Printf("Listening on unix domain socket: %s", s.SocketPath)
	} else if s.CertFile != "" && s.KeyFile != "" {
		cfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		httpServer.TLSConfig = cfg
		httpServer.TLSConfig.Certificates = make([]tls.Certificate, 1)
		httpServer.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
		httpServer.Addr = fmt.Sprintf("%s:%d", s.ListenAddress, s.Port)
		if err != nil {
			return err
		}

		s.Scheme = "https"
		log.Infof("Listening on %s://%s:%d", s.Scheme, s.ListenAddress, s.Port)
	} else {
		s.Scheme = "http"
		httpServer.Addr = fmt.Sprintf("%s:%d", s.ListenAddress, s.Port)
		log.Infof("Listening on %s://%s:%d", s.Scheme, s.ListenAddress, s.Port)
	}

	s.httpServer = httpServer
	if err := e.StartServer(httpServer); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	return s.httpServer.Shutdown(ctx)
}
