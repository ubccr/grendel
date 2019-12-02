package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/util"
)

const (
	DefaultPort = 80
)

var log = logger.GetLogger("HTTP")

type Server struct {
	ListenAddress net.IP
	ServerAddress net.IP
	Port          int
	Scheme        string
	KeyFile       string
	CertFile      string
	Hostname      string
	DB            model.Datastore
}

func NewServer(db model.Datastore, address string, port int) (*Server, error) {
	s := &Server{DB: db}

	if address == "" {
		address = net.IPv4zero.String()
	}

	if port == 0 {
		port = DefaultPort
	}

	s.Port = port

	ip := net.ParseIP(address)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("Invalid IPv4 address: %s", address)
	}

	s.ListenAddress = ip

	if !ip.To4().Equal(net.IPv4zero) {
		s.ServerAddress = ip
		return s, nil
	}

	ipaddr, err := util.GetInterfaceIP()
	if err != nil {
		return nil, err
	}

	s.ServerAddress = ipaddr

	return s, nil
}

func HTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	if code == http.StatusNotFound {
		log.WithFields(logrus.Fields{
			"path": c.Request().URL,
			"ip":   c.RealIP(),
		}).Error("Requested path not found")
	}

	c.String(code, "")
	c.Logger().Error(err)
}

func (s *Server) Serve() error {
	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Logger = EchoLogger()

	renderer, err := NewTemplateRenderer()
	if err != nil {
		return err
	}

	e.Renderer = renderer

	h, err := NewHandler(s.DB)
	if err != nil {
		return err
	}

	h.SetupRoutes(e)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.ListenAddress, s.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

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
		log.Printf("Running on https://%s:%d", s.ListenAddress, s.Port)
	} else {
		log.Warn("**WARNING*** SSL/TLS not enabled. HTTP communication will not be encrypted and vulnerable to snooping.")
		log.Printf("Running on http://%s:%d", s.ListenAddress, s.Port)
	}

	return e.StartServer(httpServer)
}
