// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/render"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/util"
)

const (
	DefaultPort = 80
)

var log = logger.GetLogger("PROVISION")

type Server struct {
	ListenAddress net.IP
	ServerAddress net.IP
	Port          int
	Scheme        string
	KeyFile       string
	CertFile      string
	RepoDir       string
	TemplatesDir  string
	DB            store.Store
	httpServer    *http.Server
}

func NewServer(db store.Store, address string) (*Server, error) {
	s := &Server{DB: db}

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

	serverAddress, err := util.ServerAddress(ip.String())
	if err != nil {
		return nil, err
	}

	s.ServerAddress = net.ParseIP(serverAddress)

	return s, nil
}

func newEcho(templates *render.Set) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Logger = EchoLogger()
	e.Renderer = &templateRenderer{set: templates}

	return e
}

// templateRenderer serves the template set through echo's c.Render
type templateRenderer struct {
	set *render.Set
}

func (t *templateRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	if viewContext, isMap := data.(map[string]any); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	ct := c.Response().Header().Get(echo.HeaderContentType)
	if ct == "" {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
	}

	return t.set.Execute(w, name, data)
}

func HTTPErrorHandler(err error, c echo.Context) {
	path := c.Request().URL.Path
	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code == http.StatusNotFound {
			log.WithFields(logrus.Fields{
				"path": path,
				"ip":   c.RealIP(),
			}).Warn("Requested path not found")
		} else {
			log.WithFields(logrus.Fields{
				"code": he.Code,
				"err":  he.Internal,
				"path": path,
				"ip":   c.RealIP(),
			}).Error(he.Message)
		}
	} else {
		log.WithFields(logrus.Fields{
			"err":  err,
			"path": path,
			"ip":   c.RealIP(),
		}).Error("HTTP Error")
	}

	c.Echo().DefaultHTTPErrorHandler(err, c)
}

func (s *Server) Serve(defaultImageName string) error {
	templates, err := render.NewSet()
	if err != nil {
		return err
	}

	if err := templates.Watch(context.Background()); err != nil {
		return err
	}

	e := newEcho(templates)

	if len(s.RepoDir) > 0 {
		log.Infof("Using repo dir: %s", s.RepoDir)
		e.Static("/repo", s.RepoDir)
	}

	if len(s.TemplatesDir) > 0 {
		log.Infof("Using templates dir: %s", s.TemplatesDir)
	}

	h, err := NewHandler(s.DB, defaultImageName)
	if err != nil {
		return err
	}

	h.Templates = templates
	h.SetupRoutes(e)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.ListenAddress, s.Port),
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	s.Scheme = "http"

	if s.CertFile != "" && s.KeyFile != "" {
		cfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
			/* TODO need to figure out compataible ciphers with iPXE

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
				tls.TLS_RSA_WITH_AES_256_CBC_SHA256,
			},
			*/
		}

		httpServer.TLSConfig = cfg
		httpServer.TLSConfig.Certificates = make([]tls.Certificate, 1)
		httpServer.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
		if err != nil {
			return err
		}

		s.Scheme = "https"
		httpServer.Addr = fmt.Sprintf("%s:%d", s.ListenAddress, s.Port)
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
