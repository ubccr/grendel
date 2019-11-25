package api

import (
	"context"
	"crypto/tls"
	"fmt"
	//	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/tftp"
	"go.universe.tf/netboot/pixiecore"
)

const (
	portDHCP = 67
	portTFTP = 69
	portHTTP = 80
	portPXE  = 4011
)

type Server struct {
	Booter pixiecore.Booter

	// Address to listen on, or empty for all interfaces.
	Address string
	// HTTP port for boot services.
	HTTPPort   int
	HTTPScheme string

	// Ipxe lists the supported bootable Firmwares, and their
	// associated ipxe binary.
	Ipxe map[pixiecore.Firmware][]byte

	// Log receives logs on Pixiecore's operation. If nil, logging
	// is suppressed.
	Log func(subsystem, msg string)
	// Debug receives extensive logging on Pixiecore's internals. Very
	// useful for debugging, but very verbose.
	Debug func(subsystem, msg string)

	// These ports can technically be set for testing, but the
	// protocols burned in firmware on the client side hardcode these,
	// so if you change them in production, nothing will work.
	DHCPPort int
	TFTPPort int
	PXEPort  int

	// Listen for DHCP traffic without binding to the DHCP port. This
	// enables coexistence of Pixiecore with another DHCP server.
	//
	// Currently only supported on Linux.
	DHCPNoBind bool

	// Private key file
	KeyFile string

	// Certificate file
	CertFile string

	// Hostname
	Hostname string

	errs     chan error
	eventsMu sync.Mutex
	events   map[string][]machineEvent
}

func HTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	if code == http.StatusNotFound {
		log.WithFields(log.Fields{
			"path": c.Request().URL,
			"ip":   c.RealIP(),
		}).Error("Requested path not found")
	}

	c.String(code, "")
	c.Logger().Error(err)
}

func (s *Server) Serve() error {
	if s.DHCPPort == 0 {
		s.DHCPPort = portDHCP
	}
	if s.TFTPPort == 0 {
		s.TFTPPort = portTFTP
	}
	if s.PXEPort == 0 {
		s.PXEPort = portPXE
	}
	if s.HTTPPort == 0 {
		s.HTTPPort = portHTTP
	}
	if s.HTTPScheme == "" {
		s.HTTPScheme = "http"
	}

	//pxeConn, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", s.Address, s.PXEPort))
	//if err != nil {
	//		return err
	//	}

	echox, httpServer, err := s.serveHTTP()
	if err != nil {
		//		pxeConn.Close()
		return err
	}

	tftpServer, err := tftp.NewServer(s.Address)
	if err != nil {
		return err
	}

	s.events = make(map[string][]machineEvent)
	// 5 buffer slots, one for each goroutine, plus one for
	// Shutdown(). We only ever pull the first error out, but shutdown
	// will likely generate some spurious errors from the other
	// goroutines, and we want them to be able to dump them without
	// blocking.
	s.errs = make(chan error, 5)

	s.debug("Init", "Starting Pixiecore goroutines")

	//	go func() { s.errs <- s.servePXE(pxeConn) }()
	go func() { s.errs <- tftpServer.Serve() }()
	go func() { s.errs <- echox.StartServer(httpServer) }()

	// Wait for either a fatal error, or Shutdown().
	err = <-s.errs
	//	pxeConn.Close()
	tftpServer.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := echox.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Failed to stop http server")
	}
	return err
}

// Start web server
func (s *Server) serveHTTP() (*echo.Echo, *http.Server, error) {
	e := echo.New()
	//e.HTTPErrorHandler = HTTPErrorHandler
	e.HideBanner = true
	e.Use(middleware.Recover())

	h, err := NewHandler(s.Booter)
	if err != nil {
		return nil, nil, err
	}

	h.SetupRoutes(e)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Address, s.HTTPPort),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if s.CertFile != "" && s.KeyFile != "" {
		cfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
			/*
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
			return nil, nil, err
		}

		s.HTTPPort = 443
		s.HTTPScheme = "https"
		httpServer.Addr = fmt.Sprintf("%s:%d", s.Address, s.HTTPPort)
		log.Printf("Running on https://%s:%d", s.Address, s.HTTPPort)
	} else {
		log.Warn("**WARNING*** SSL/TLS not enabled. HTTP communication will not be encrypted and vulnerable to snooping.")
		log.Printf("Running on http://%s:%d", s.Address, s.HTTPPort)
	}

	return e, httpServer, nil
}
