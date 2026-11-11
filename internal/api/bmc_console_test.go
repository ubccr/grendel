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
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func TestBmcConsole(t *testing.T) {
	h := newTestHandler(t)
	h.Redfish = redfish.New(redfish.Config{})

	host := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHost(host); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(h.BmcConsole))
	t.Cleanup(srv.Close)

	t.Run("refused before the upgrade", func(t *testing.T) {
		for query, want := range map[string]int{
			"":                           http.StatusBadRequest,
			"?node=notindb01":            http.StatusNotFound,
			"?node=" + host.Name + ",x1": http.StatusNotFound,
		} {
			res, err := http.Get(srv.URL + query)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()
			if res.StatusCode != want {
				t.Errorf("%q: status %d, want %d", query, res.StatusCode, want)
			}
		}
	})

	t.Run("a console that can't open closes with the reason", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"?node="+host.Name, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.CloseNow()

		_, _, err = conn.Read(ctx)
		if websocket.CloseStatus(err) != ConsoleFailed {
			t.Fatalf("close status = %v, want %d", err, ConsoleFailed)
		}
		if !strings.Contains(err.Error(), "no bmc credentials configured") {
			t.Errorf("close reason = %v, want the console error", err)
		}
	})
}
