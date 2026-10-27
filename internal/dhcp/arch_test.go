// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package dhcp

import (
	"testing"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/iana"
	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/pkg/model"
)

func TestDetectArch(t *testing.T) {
	assert := assert.New(t)

	host := &model.Host{Name: "cpn-01"}

	req, err := dhcpv4.New(dhcpv4.WithOption(dhcpv4.OptClientArch(iana.EFI_ARM64)))
	if assert.NoError(err) {
		assert.Equal("aarch64", detectArch(host, req))
	}

	// requests without option 93, or with an unsupported arch, detect nothing
	req, err = dhcpv4.New()
	if assert.NoError(err) {
		assert.Equal("", detectArch(host, req))
	}

	req, err = dhcpv4.New(dhcpv4.WithOption(dhcpv4.OptClientArch(iana.EFI_ARM32)))
	if assert.NoError(err) {
		assert.Equal("", detectArch(host, req))
	}
}
