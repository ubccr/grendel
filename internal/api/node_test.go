// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
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
