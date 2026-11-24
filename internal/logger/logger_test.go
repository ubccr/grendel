// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package logger

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/settings/settingstest"
)

func TestEachLoggerWritesAtItsOwnLevel(t *testing.T) {
	var out bytes.Buffer
	shared := GetLogger("").Logger
	before := shared.Out
	shared.SetOutput(&out)
	t.Cleanup(func() {
		shared.SetOutput(before)
		shared.SetLevel(logrus.InfoLevel)
	})

	settingstest.Set(t, "loggers.dns", "debug")
	settingstest.Set(t, "loggers.tftp", "off")
	Apply()

	GetLogger("DNS").Debug("dns detail")
	GetLogger("DHCP").Debug("dhcp detail")
	GetLogger("DHCP").Info("dhcp routine")
	GetLogger("DHCP").Warn("dhcp warning")
	GetLogger("TFTP").Error("tftp error")
	GetLogger("unlisted").Warn("unlisted warning")

	assert.Contains(t, out.String(), "dns detail")
	assert.Contains(t, out.String(), "dhcp warning")
	assert.Contains(t, out.String(), "unlisted warning")
	assert.NotContains(t, out.String(), "dhcp detail", "one logger at debug does not turn it on for the rest")
	assert.NotContains(t, out.String(), "dhcp routine")
	assert.NotContains(t, out.String(), "tftp error")

	out.Reset()
	settingstest.Set(t, "log_level", "info")
	Apply()
	GetLogger("DHCP").Info("dhcp routine")
	assert.Contains(t, out.String(), "dhcp routine")
}
