// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/go-faster/jx"
	"github.com/ubccr/grendel/pkg/client"
)

// TestRedfishNullsAreNullable checks the committed spec against the fields reflection says can arrive as null. A field missing its nullable flag makes the generated client reject the response, so this fails either when a gofish upgrade adds a field or when someone edits the types without regenerating the spec.
func TestRedfishNullsAreNullable(t *testing.T) {
	doc := loadSpec(t)

	missing := notNullable(doc, "")
	for _, path := range missing {
		t.Errorf("%s can be null but the spec does not say so, regenerate api/openapi.json", path)
	}
}

// notNullable returns the path of every property named in redfishNulls that the spec does not mark nullable.
func notNullable(node any, path string) []string {
	var found []string

	switch n := node.(type) {
	case map[string]any:
		for key, sub := range n {
			schema, ok := sub.(map[string]any)
			if ok && slices.Contains(redfishNulls, key) && schema["nullable"] != true {
				found = append(found, path+"."+key)
			}
			found = append(found, notNullable(sub, path+"."+key)...)
		}
	case []any:
		for i, sub := range n {
			found = append(found, notNullable(sub, fmt.Sprintf("%s[%d]", path, i))...)
		}
	}

	return found
}

// TestDecodeJobWithNulls is the payload an iDRAC returns for a job that carries no parameters and whose message has no arguments or resolution steps.
func TestDecodeJobWithNulls(t *testing.T) {
	const payload = `{
		"name": "cpn-42-bmc",
		"jobs": [{
			"Id": "JID_123456789",
			"JobState": "Completed",
			"Oem": null,
			"Parameters": null,
			"PercentComplete": 100,
			"RawData": null,
			"StepOrder": null,
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
