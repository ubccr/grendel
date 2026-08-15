// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"testing"

	"github.com/ubccr/grendel/internal/store/sqlstore"
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
