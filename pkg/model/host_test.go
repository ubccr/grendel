// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func TestHostTags(t *testing.T) {
	assert := assert.New(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.Tags = []string{"k11", "switch", "dellztd"}

	assert.True(host.HasTags("k11", "dellztd"))
	assert.False(host.HasTags("p22", "dellztd"))
	assert.True(host.HasAnyTags("p22", "dellztd"))
	assert.False(host.HasAnyTags("p22", "m12"))
	assert.False(host.HasAnyTags())
	assert.False(host.HasTags())
	assert.Equal("", host.Interfaces[0].HostNameIndex(100))
	assert.Equal(host.Interfaces[0].FQDN, host.Interfaces[0].HostNameIndex(0))
}

func TestHostBonds(t *testing.T) {
	assert := assert.New(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	assert.Equal(host.Bonds[0].AddrString(), host.Bonds[0].IP.Addr().String())
}

func BenchmarkGJSONUnmarshall(b *testing.B) {
	jsonStr := string(tests.TestHostJSON)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		host := &model.Host{}
		host.FromJSON(jsonStr)
	}
}

func BenchmarkGJSONMarshall(b *testing.B) {
	jsonStr := string(tests.TestHostJSON)
	host := &model.Host{}
	host.FromJSON(jsonStr)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		host.ToJSON()
	}
}

func BenchmarkEncodeUnmarshall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var host model.Host
		err := json.Unmarshal(tests.TestHostJSON, &host)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeMarshall(b *testing.B) {
	var host model.Host
	err := json.Unmarshal(tests.TestHostJSON, &host)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := json.Marshal(&host)
		if err != nil {
			b.Fatal(err)
		}
	}
}
