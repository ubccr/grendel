// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/go-faster/jx"
	"github.com/ubccr/grendel/pkg/client"
)

// TestRedfishArraysAreNullable walks the committed spec for arrays the schema customizer has not marked nullable. gofish tags none of its slices omitempty, so an empty one marshals as null and the generated client refuses to decode it. The browser client does not validate, so a miss here breaks only the CLI
func TestRedfishArraysAreNullable(t *testing.T) {
	doc := loadSpec(t)

	schemas, ok := doc["components"].(map[string]any)["schemas"].(map[string]any)
	if !ok {
		t.Fatal("spec has no component schemas")
	}

	for name, schema := range schemas {
		if !strings.HasPrefix(name, "Redfish") {
			continue
		}
		for _, path := range nonNullableArrays(schema, name) {
			t.Errorf("%s is an array that is not nullable, add its field name to redfishNulls in options.go and regenerate", path)
		}
	}
}

// TestDecodeJobWithNullArrays is the payload an iDRAC returns for a job whose messages carry no arguments or resolution steps.
func TestDecodeJobWithNullArrays(t *testing.T) {
	const payload = `{
		"name": "cpn-42-bmc",
		"jobs": [{
			"Id": "JID_123456789",
			"JobState": "Completed",
			"PercentComplete": 100,
			"Messages": [{
				"Message": "The command was successful",
				"MessageArgs": null,
				"MessageId": "SUP019",
				"RelatedProperties": null,
				"ResolutionSteps": null
			}]
		}]
	}`

	var job client.RedfishJob
	if err := job.Decode(jx.DecodeStr(payload)); err != nil {
		t.Fatalf("decoding a job with null arrays: %v", err)
	}

	msgs := job.Jobs.Value[0].Value.Messages
	if got := len(msgs.Value); got != 1 {
		t.Fatalf("got %d messages, want 1", got)
	}
	if got := msgs.Value[0].Value.Message.Value; got != "The command was successful" {
		t.Errorf("got message %q", got)
	}
	if steps := msgs.Value[0].Value.ResolutionSteps; !steps.Null {
		t.Errorf("ResolutionSteps decoded as %+v, want null", steps)
	}
}

func loadSpec(t *testing.T) map[string]any {
	t.Helper()

	f, err := os.Open("../../api/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var doc map[string]any
	if err := json.NewDecoder(f).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

// nonNullableArrays returns the path of every array in the schema that may still be handed a null.
func nonNullableArrays(node any, path string) []string {
	schema, ok := node.(map[string]any)
	if !ok {
		return nil
	}

	var found []string
	if schema["type"] == "array" && schema["nullable"] != true {
		found = append(found, path)
	}

	if props, ok := schema["properties"].(map[string]any); ok {
		for name, sub := range props {
			found = append(found, nonNullableArrays(sub, path+"."+name)...)
		}
	}
	found = append(found, nonNullableArrays(schema["items"], path+"[]")...)

	return found
}
