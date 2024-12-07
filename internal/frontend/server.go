// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frontend

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/memory"
	"github.com/gofiber/storage/redis"
	"github.com/gofiber/storage/sqlite3"
	"github.com/gofiber/template/html/v2"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/util"
)

const (
	DefaultPort = 8081
)

var log = logger.GetLogger("Frontend")

//go:embed views
var staticFS embed.FS

type Server struct {
	ListenAddress net.IP
	ServerAddress net.IP
	Port          int
	Scheme        string
	KeyFile       string
	CertFile      string
	DB            store.Store
	app           *fiber.App
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

func newStore() fiber.Storage {
	var storage fiber.Storage

	switch viper.GetString("frontend.session_storage") {
	case "redis":
		storage = redis.New(redis.Config{
			URL:   viper.GetString("frontend.redis.url"),
			Reset: false,
		})
	case "sqlite3":
		storage = sqlite3.New(sqlite3.Config{
			Database: viper.GetString("frontend.sqlite3.path"),
			Table:    "grendel_frontend_sessions",
		})
	default:
		storage = memory.New()
	}

	return storage
}

func (s *Server) Serve() error {
	storage := newStore()
	store := session.New(session.Config{
		Expiration:     8 * time.Hour,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		Storage:        storage,
	})

	h, err := NewHandler(s.DB, store)
	if err != nil {
		return err
	}
	var funcMap = fiber.Map{
		"Split":   strings.Split,
		"Join":    strings.Join,
		"Sprintf": fmt.Sprintf,
		"Iterate": func(count int) []string {
			var Items []string
			for i := 0; i < count; i++ {
				Items = append(Items, fmt.Sprint(i))
			}
			return Items
		},
		"StrAdd": func(x, y string) int {
			a, _ := strconv.Atoi(x)
			b, _ := strconv.Atoi(y)
			return a + b
		},
	}

	views, err := fs.Sub(staticFS, "views")
	if err != nil {
		log.Error("failed to load views")
		return err
	}
	engine := html.NewFileSystem(http.FS(views), ".gohtml")
	engine.AddFuncMap(funcMap)
	s.app = fiber.New(fiber.Config{
		Views:                 engine,
		ViewsLayout:           "base",
		DisableStartupMessage: true,
	})

	h.SetupRoutes(s.app)

	if s.CertFile != "" && s.KeyFile != "" {
		s.Scheme = "https"
		err = s.app.ListenTLS(fmt.Sprintf("%s:%d", s.ListenAddress, s.Port), s.CertFile, s.KeyFile)
	} else {
		s.Scheme = "http"
		err = s.app.Listen(fmt.Sprintf("%s:%d", s.ListenAddress, s.Port))
	}

	log.Infof("Listening on %s://%s:%d", s.Scheme, s.ListenAddress, s.Port)
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.app == nil {
		return nil
	}

	return s.app.Shutdown()
}
