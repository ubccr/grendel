// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

// a shutdown can land before Serve has assigned the fuego server, when grendel is interrupted while it is still starting up. That is a clean stop with nothing to do, not a failure, and reporting it as one exits grendel non-zero on an ordinary ctrl-c
func TestShutdownBeforeServe(t *testing.T) {
	db, err := sqlstore.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	for _, tt := range []struct {
		name    string
		socket  string
		address string
	}{
		{"socket only", "grendel-test.socket", ""},
		{"tcp only", "", "127.0.0.1:0"},
		{"socket and tcp", "grendel-test.socket", "127.0.0.1:0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewServer(db, tt.socket, tt.address)
			if err != nil {
				t.Fatal(err)
			}

			if err := s.Shutdown(context.Background()); err != nil {
				t.Errorf("Shutdown before Serve returned an error: %s", err)
			}
		})
	}
}

func newWiredTestServer(t *testing.T) (*sqlstore.SqlStore, *fuego.Server) {
	t.Helper()
	viper.Set("api.secret", "wired-test-secret-0123456789abcdef")
	t.Cleanup(func() { viper.Set("api.secret", nil) })

	db, err := sqlstore.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	h, err := NewHandler(db)
	require.NoError(t, err)
	h.Redfish = redfish.New(redfish.Config{})

	s := (&Server{DB: db}).newFuego()
	h.SetupRoutes(s)
	return db, s
}

// a bmc run that outlasts the server's write timeout has its response cut off unless the handler can clear the deadline, which a middleware wrapping the response writer silently prevents
func TestHandlersCanClearTheWriteDeadline(t *testing.T) {
	_, s := newWiredTestServer(t)

	var cleared error
	fuego.Get(s, "/deadline", func(c fuego.ContextNoBody) (string, error) {
		cleared = http.NewResponseController(c.Response()).SetWriteDeadline(time.Time{})
		return "", nil
	})

	srv := httptest.NewServer(s.Mux)
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/deadline")
	require.NoError(t, err)
	res.Body.Close()
	require.NoError(t, cleared)
}

func TestBmcConsoleRoute(t *testing.T) {
	db, s := newWiredTestServer(t)
	session := signinAs(t, db, s.Mux, "admin", model.RoleAdmin.String())

	host := tests.HostFactory.MustCreate().(*model.Host)
	require.NoError(t, db.StoreHost(host))

	srv := httptest.NewServer(s.Mux)
	t.Cleanup(srv.Close)
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/v1/bmc/console?node=" + host.Name

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("without a session it is refused before the upgrade", func(t *testing.T) {
		_, res, err := websocket.Dial(ctx, url, nil)
		require.Error(t, err)
		require.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("with a session it upgrades", func(t *testing.T) {
		header := http.Header{}
		for _, c := range session {
			header.Add("Cookie", c.Name+"="+c.Value)
		}
		conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: header})
		require.NoError(t, err)
		defer conn.CloseNow()

		_, _, err = conn.Read(ctx)
		require.Equal(t, ConsoleFailed, websocket.CloseStatus(err), "the console has no credentials, so it closes with the reason: %v", err)
	})
}
