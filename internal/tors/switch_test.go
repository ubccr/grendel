// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"errors"
	"fmt"
	"net/netip"
	"testing"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/pkg/model"
)

func testConfig() Config {
	return Config{
		Username:  "admin",
		Password:  "secret",
		GNMI:      GNMIConfig{Port: 6030, NetworkInstance: "default"},
		SNMP:      SNMPConfig{Version: "2c", Community: "public", Port: 161},
		gnmiByNOS: map[string]GNMIConfig{NOSSonic: {Port: 8080, NetworkInstance: "default"}},
	}
}

func TestNewUsesFirstInterface(t *testing.T) {
	first := &model.NetInterface{IP: netip.MustParsePrefix("10.0.0.1/24")}
	second := &model.NetInterface{IP: netip.MustParsePrefix("10.0.0.2/24")}

	tests := []struct {
		name       string
		interfaces []*model.NetInterface
		wantErr    error
	}{
		{name: "a switch that is not flagged bmc", interfaces: []*model.NetInterface{first, second}},
		{name: "the flag does not move it to another interface", interfaces: []*model.NetInterface{first, {IP: second.IP, BMC: true}}},
		{name: "the flag on the first interface changes nothing", interfaces: []*model.NetInterface{{IP: first.IP, BMC: true}, second}},
		{name: "no interfaces is an error", wantErr: ErrNoInterfaces},
		{name: "a first interface with no IP is an error, not a dial to this host", interfaces: []*model.NetInterface{{}, second}, wantErr: ErrNoInterfaces},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := &model.Host{Name: "swe-a01-01", Tags: []string{"sonic"}, Interfaces: tt.interfaces}

			sw, err := New(host, testConfig())
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if got := sw.(*GNMI).target(); got != "10.0.0.1:8080" {
				t.Errorf("target = %q, want the first interface on the sonic port", got)
			}
		})
	}
}

func TestNewPicksDriverFromTag(t *testing.T) {
	iface := []*model.NetInterface{{IP: netip.MustParsePrefix("10.0.0.1/24")}}
	tests := []struct {
		tag     string
		check   func(NetworkSwitch) bool
		wantErr error
	}{
		{tag: "arista", check: func(s NetworkSwitch) bool { g, ok := s.(*GNMI); return ok && g.target() == "10.0.0.1:6030" }},
		{tag: "sonic", check: func(s NetworkSwitch) bool { _, ok := s.(*GNMI); return ok }},
		{tag: "dellos10", check: func(s NetworkSwitch) bool { _, ok := s.(*SNMP); return ok }},
		{tag: "cumulus", wantErr: ErrUnknownNOS},
	}
	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			sw, err := New(&model.Host{Tags: []string{"switch", tt.tag}, Interfaces: iface}, testConfig())
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("err = %v, want exactly %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !tt.check(sw) {
				t.Errorf("wrong driver %T", sw)
			}
		})
	}
}

func TestGNMINeedsCredentialsButSNMPDoesNot(t *testing.T) {
	cfg := testConfig()
	cfg.Username, cfg.Password = "", ""
	if _, err := NewFromAddr("10.0.0.1", NOSArista, cfg); !errors.Is(err, ErrNoCredentials) {
		t.Errorf("arista err = %v, want ErrNoCredentials", err)
	}
	if _, err := NewFromAddr("10.0.0.1", NOSDellOS10, cfg); err != nil {
		t.Errorf("dellos10 err = %v, want nil", err)
	}
}

func TestConfigFromViper(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set("bmc.switch_admin_username", "legacy")
	viper.Set("bmc.switch_admin_password", "legacy-pw")
	viper.Set("switch.gnmi.port", 6030)
	viper.Set("switch.gnmi.sonic.port", 8080)
	viper.Set("switch.gnmi.sonic.plaintext", false)
	viper.Set("switch.gnmi.plaintext", true)

	cfg := ConfigFromViper()
	if cfg.Username != "legacy" || cfg.Password != "legacy-pw" {
		t.Errorf("legacy bmc.switch_admin_* not used: %q/%q", cfg.Username, cfg.Password)
	}
	if g := cfg.gnmiFor(NOSArista); g.Port != 6030 || !g.Plaintext {
		t.Errorf("arista = %+v, want the global settings", g)
	}
	if g := cfg.gnmiFor(NOSSonic); g.Port != 8080 || g.Plaintext {
		t.Errorf("sonic = %+v, want its overrides", g)
	}

	viper.Set("switch.username", "only-user")
	if c := ConfigFromViper(); c.Username != "only-user" || c.Password != "legacy-pw" {
		t.Errorf("a username alone should still take the legacy password: %q/%q", c.Username, c.Password)
	}
	viper.Set("switch.username", "")

	if cfg.Timeout != defaultConfig.Timeout || cfg.SNMP.Port != 161 || !cfg.gnmiFor(NOSSonic).SkipVerify {
		t.Errorf("defaults lost once [switch] has keys: %+v", cfg)
	}

	viper.Set("switch.username", "new")
	viper.Set("switch.password", "new-pw")
	if cfg := ConfigFromViper(); cfg.Username != "new" {
		t.Errorf("switch.username = %q, want it to win over the legacy keys", cfg.Username)
	}
}

func TestErrorCode(t *testing.T) {
	cases := map[error]string{
		ErrUnknownNOS:                            CodeUnknownNOS,
		fmt.Errorf("%w: %q", ErrUnknownNOS, "x"): CodeUnknownNOS,
		ErrNoCredentials:                         CodeNoCredentials,
		ErrNoInterfaces:                          CodeNoInterfaces,
		errors.New("i/o timeout"):                CodeUnreachable,
	}
	for err, want := range cases {
		if got := ErrorCode(err); got != want {
			t.Errorf("ErrorCode(%v) = %q, want %q", err, got, want)
		}
	}
}
