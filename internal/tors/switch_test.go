// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"net/netip"
	"testing"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/pkg/model"
)

func TestNewNetworkSwitchUsesFirstInterface(t *testing.T) {
	viper.Set("bmc.switch_admin_username", "admin")
	viper.Set("bmc.switch_admin_password", "secret")
	t.Cleanup(func() {
		viper.Set("bmc.switch_admin_username", "")
		viper.Set("bmc.switch_admin_password", "")
	})

	first := &model.NetInterface{IP: netip.MustParsePrefix("10.0.0.1/24")}
	second := &model.NetInterface{IP: netip.MustParsePrefix("10.0.0.2/24")}

	tests := []struct {
		name       string
		interfaces []*model.NetInterface
		wantErr    bool
	}{
		{name: "a switch that is not flagged bmc", interfaces: []*model.NetInterface{first, second}},
		{name: "the flag does not move it to another interface", interfaces: []*model.NetInterface{first, {IP: second.IP, BMC: true}}},
		{name: "the flag on the first interface changes nothing", interfaces: []*model.NetInterface{{IP: first.IP, BMC: true}, second}},
		{name: "no interfaces is an error", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := &model.Host{Name: "swe-a01-01", Tags: []string{"sonic"}, Interfaces: tt.interfaces}

			sw, err := NewNetworkSwitch(host)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if got := sw.(*Sonic).baseUrl; got != "https://10.0.0.1" {
				t.Errorf("baseUrl = %q, want the first interface", got)
			}
		})
	}
}
