package frontend

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/memory"
	"github.com/gofiber/storage/redis"
	"github.com/gofiber/storage/sqlite3"
	"github.com/gofiber/template/html/v2"
	"github.com/spf13/viper"
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
	KeyFile       string
	CertFile      string
	DB            model.DataStore
	app           *fiber.App
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
	}

	engine := html.New("./frontend/views/", ".gohtml")
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
	s.app.Use(cors.New(cors.Config{
		// TODO: verify this is correct
		AllowOrigins:     fmt.Sprintf("%s://%s:%d", s.Scheme, s.ServerAddress, s.Port),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Cache-Control",
		AllowCredentials: true,
	}))

	log.Infof("Listening on %s://%s:%d", s.Scheme, s.ListenAddress, s.Port)
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.app == nil {
		return nil
	}

	return s.app.Shutdown()
}
