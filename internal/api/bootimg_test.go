// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/pkg/model"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()

	db, err := sqlstore.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	h, err := NewHandler(db)
	if err != nil {
		t.Fatal(err)
	}

	images := model.BootImageList{
		&model.BootImage{Name: "image1", KernelPath: "/tmp/image1/vmlinuz"},
		&model.BootImage{Name: "image2", KernelPath: "/tmp/image2/vmlinuz"},
	}
	if err := db.StoreBootImages(images); err != nil {
		t.Fatal(err)
	}

	return h
}

func TestBootImageFind(t *testing.T) {
	h := newTestHandler(t)

	tests := []struct {
		name  string
		names string
		unset bool
		want  []string
	}{
		{name: "empty names returns all", names: "", want: []string{"image1", "image2"}},
		{name: "unset names returns all", unset: true, want: []string{"image1", "image2"}},
		{name: "single name", names: "image2", want: []string{"image2"}},
		{name: "multiple names", names: "image1,image2", want: []string{"image1", "image2"}},
		{name: "unknown name", names: "nope", want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fuego.NewMockContextNoBody()
			if !tt.unset {
				c.SetQueryParam("names", tt.names)
			}

			res, err := h.BootImageFind(c)
			if err != nil {
				t.Fatal(err)
			}

			if len(res) != len(tt.want) {
				t.Fatalf("got %d images, want %d", len(res), len(tt.want))
			}

			for _, name := range tt.want {
				found := false
				for _, image := range res {
					if image.Name == name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing image %s in results", name)
				}
			}
		})
	}
}
