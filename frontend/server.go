package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/util"
)

const (
	DefaultPort = 8080
)

var log = logger.GetLogger("Frontend")

type Server struct {
	ListenAddress net.IP
	ServerAddress net.IP
	Port          int
	Scheme        string
	// KeyFile       string
	// CertFile      string
	// RepoDir       string
	DB         model.DataStore
	httpServer *http.Server
}

func NewServer(db model.DataStore, address string) (*Server, error) {
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
		return nil, fmt.Errorf("invalid ipv4 address: %s", shost)
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
	h, err := NewHandler(s.DB)
	if err != nil {
		return err
	}
	var funcMap = fiber.Map{
		"Split":     strings.Split,
		"Join":      strings.Join,
		"Sprintf":   fmt.Sprintf,
		"Stringify": func(v interface{}) string {
			b, _ := json.MarshalIndent(v, "", "    ")
			return string(b)
		},
	}

	engine := html.New("./frontend/views/", ".gohtml")
	engine.AddFuncMap(funcMap)

	app := fiber.New(fiber.Config{
		Views: engine,
		ViewsLayout: "base",
	})

	h.SetupRoutes(app)

	app.Listen(fmt.Sprintf("%s:%d", s.ListenAddress, s.Port))

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	return s.httpServer.Shutdown(ctx)
}
