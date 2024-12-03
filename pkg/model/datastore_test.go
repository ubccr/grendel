// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model_test

import (
	"encoding/json"
	"testing"

	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

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
