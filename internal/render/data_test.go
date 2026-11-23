// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package render

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/pkg/model"
)

func TestBeaconTimeout(t *testing.T) {
	now := time.Now()
	token, err := model.NewBootToken("2BH7b1vYyOQgC2pzGJ9lOqSHvLq", "aa:bb:cc:dd:ee:ff")
	require.NoError(t, err)

	expires, err := model.BootTokenExpiry(token)
	require.NoError(t, err)
	assert.WithinDuration(t, now.Add(time.Hour), expires, 2*time.Second)

	// the beacon gives up ten minutes before its token stops being accepted
	got, err := time.ParseDuration(beaconTimeout(token, now))
	require.NoError(t, err)
	assert.InDelta(t, (50 * time.Minute).Seconds(), got.Seconds(), 2)

	// a token near its end keeps half of what is left for the report
	got, err = time.ParseDuration(beaconTimeout(token, expires.Add(-4*time.Minute)))
	require.NoError(t, err)
	assert.Equal(t, 2*time.Minute, got)

	assert.Empty(t, beaconTimeout(token, expires.Add(time.Minute)), "an expired token")
	assert.Empty(t, beaconTimeout("not a token!", now))

	viper.Set("provision.token_ttl", 0)
	t.Cleanup(func() { viper.Set("provision.token_ttl", 60*60) })
	assert.Empty(t, beaconTimeout(token, now), "tokens that never expire leave the beacon its own default")
}

func TestRenderString(t *testing.T) {
	assert := assert.New(t)

	host := &model.Host{Name: "cpn-01", Tags: []string{"gpu"}}
	image := &model.BootImage{Name: "ubuntu", Version: "24.04.1"}
	data := map[string]any{"host": host, "bootimage": image}

	tests := []struct {
		text    string
		want    string
		wantErr bool
	}{
		{text: "/var/lib/grendel/images/vmlinuz", want: "/var/lib/grendel/images/vmlinuz"},
		{text: "/images/{{ .host.Name }}/vmlinuz", want: "/images/cpn-01/vmlinuz"},
		{text: `/images/{{ if hasTag .host "gpu" }}gpu{{ else }}cpu{{ end }}/vmlinuz`, want: "/images/gpu/vmlinuz"},
		{text: "/images/{{ .bootimage.Name }}-{{ .bootimage.Version }}/vmlinuz", want: "/images/ubuntu-24.04.1/vmlinuz"},
		{text: "/images/{{ .nic.Name }}/vmlinuz", wantErr: true},
		{text: "/images/{{ .host.Name /vmlinuz", wantErr: true},
	}

	for _, test := range tests {
		got, err := RenderString("test", test.text, data)
		if test.wantErr {
			assert.Errorf(err, "expected error for %q", test.text)
			continue
		}
		if assert.NoErrorf(err, "unexpected error for %q", test.text) {
			assert.Equal(test.want, got)
		}
	}
}
