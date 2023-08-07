package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
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

// init http server (e)
func newEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Logger = EchoLogger()
	e.Validator = &CustomValidator{validator: validator.New()}

	return e
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

type Template struct {
	templates map[string]*template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates[name].ExecuteTemplate(w, "base.html", data)
}

func (s *Server) Serve() error {
	h, err := NewHandler(s.DB)
	if err != nil {
		return err
	}
	var funcMap = template.FuncMap{
		"Split":     strings.Split,
		"Join":      strings.Join,
		"Sprintf":   fmt.Sprintf,
		"Provision": h.DB.ProvisionHosts,
		"Stringify": func(v interface{}) string {
			b, _ := json.MarshalIndent(v, "", "    ")
			return string(b)
		},
	}
	templates := make(map[string]*template.Template)
	templates["index.html"] = template.Must(template.ParseFiles("frontend/views/index.html", "frontend/views/base.html"))
	templates["host.html"] = template.Must(template.New("host.html").Funcs(funcMap).ParseFiles("frontend/views/host.html", "frontend/views/base.html"))
	templates["floorplan.html"] = template.Must(template.New("floorplan.html").Funcs(funcMap).ParseFiles("frontend/views/floorplan.html", "frontend/views/base.html"))
	templates["rack.html"] = template.Must(template.New("rack.html").Funcs(funcMap).ParseFiles("frontend/views/rack.html", "frontend/views/base.html"))

	e := newEcho()
	e.Renderer = &Template{
		templates: templates,
	}

	h.SetupRoutes(e)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.ListenAddress, s.Port),
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	s.Scheme = "http"
	s.httpServer = httpServer
	log.Infof("Listening on %s://%s:%d", s.Scheme, s.ListenAddress, s.Port)
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
