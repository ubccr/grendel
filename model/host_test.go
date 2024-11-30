// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/model"
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
