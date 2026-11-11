// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"slices"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

func TestNodeBootTokenArch(t *testing.T) {
	h := newTestHandler(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHost(host); err != nil {
		t.Fatal(err)
	}

	for _, arch := range []string{"", "aarch64"} {
		c := fuego.NewMockContextNoBody()
		c.PathParams["interface"] = "boot"
		c.SetQueryParam("nodeset", host.Name)
		c.SetQueryParam("arch", arch)

		res, err := h.NodeBootToken(c)
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Nodes) != 1 {
			t.Fatalf("got %d tokens, want 1", len(res.Nodes))
		}

		claims, err := model.ParseBootToken(res.Nodes[0].Token)
		if err != nil {
			t.Fatal(err)
		}
		if claims.ID != host.UID.String() {
			t.Errorf("got token for %s, want %s", claims.ID, host.UID.String())
		}
		if claims.Arch != arch {
			t.Errorf("got arch %q, want %q", claims.Arch, arch)
		}
	}
}

func TestNodeBootTokenInterface(t *testing.T) {
	h := newTestHandler(t)

	withBMC := tests.HostFactory.MustCreate().(*model.Host)
	noBMC := tests.HostFactory.MustCreate().(*model.Host)
	for _, nic := range noBMC.Interfaces {
		nic.BMC = false
	}
	if err := h.DB.StoreHosts(model.HostList{withBMC, noBMC}); err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"boot": withBMC.BootInterface().MAC.String(),
		"bmc":  withBMC.InterfaceBMC().MAC.String(),
	}

	for iface, want := range tests {
		c := fuego.NewMockContextNoBody()
		c.PathParams["interface"] = iface
		c.SetQueryParam("nodeset", withBMC.Name+","+noBMC.Name)

		res, err := h.NodeBootToken(c)
		if err != nil {
			t.Fatal(err)
		}

		for _, node := range res.Nodes {
			if node.Name != withBMC.Name {
				if iface == "bmc" {
					t.Errorf("got bmc token for %s which has no bmc interface", node.Name)
				}
				continue
			}

			claims, err := model.ParseBootToken(node.Token)
			if err != nil {
				t.Fatal(err)
			}
			if claims.MAC != want {
				t.Errorf("%s token has mac %s, want %s", iface, claims.MAC, want)
			}
		}
	}
}

func TestNodeFind(t *testing.T) {
	h := newTestHandler(t)

	tagged := tests.HostFactory.MustCreate().(*model.Host)
	untagged := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHosts(model.HostList{tagged, untagged}); err != nil {
		t.Fatal(err)
	}

	ns, err := nodeset.NewNodeSet(tagged.Name)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.DB.TagHosts(ns, []string{"gpu"}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		nodeset string
		tags    string
		want    []string
		wantErr bool
	}{
		{name: "no filter returns every node", want: []string{tagged.Name, untagged.Name}},
		{name: "nodeset", nodeset: untagged.Name, want: []string{untagged.Name}},
		{name: "tags", tags: "gpu", want: []string{tagged.Name}},
		{name: "nodeset and tags intersect", nodeset: tagged.Name + "," + untagged.Name, tags: "gpu", want: []string{tagged.Name}},
		{name: "a filter that matches nothing returns nothing rather than every node", nodeset: untagged.Name, tags: "gpu"},
		{name: "a name not in the database returns nothing", nodeset: "notindb01"},
		{name: "an unknown tag is an error", tags: "typo", wantErr: true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c := fuego.NewMockContextNoBody()
			c.SetQueryParam("nodeset", tt.nodeset)
			c.SetQueryParam("tags", tt.tags)

			hosts, err := h.NodeFind(c)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}

			got := make([]string, 0, len(hosts))
			for _, host := range hosts {
				got = append(got, host.Name)
			}
			slices.Sort(got)

			want := slices.Clone(tt.want)
			slices.Sort(want)

			if !slices.Equal(got, want) {
				t.Errorf("hosts = %v, want %v", got, want)
			}
		})
	}
}
