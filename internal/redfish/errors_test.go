// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"syscall"
	"testing"

	"github.com/stmcginnis/gofish/schemas"
	"github.com/stretchr/testify/assert"
)

// a real iDRAC body, where the top level message is useless and the extended info carries the actual problem
const dellGeneralError = `{
  "error": {
    "code": "Base.1.12.GeneralError",
    "message": "A general error has occurred. See ExtendedInfo for more information",
    "@Message.ExtendedInfo": [
      {
        "Message": "Unable to complete the operation because the JobID entered is invalid.",
        "MessageArgs": [],
        "MessageId": "IDRAC.2.9.SUP020",
        "RelatedProperties": [],
        "Resolution": "Enter a valid JobID and retry the operation.",
        "Severity": "Warning"
      }
    ]
  }
}`

const dellAuthError = `{
  "error": {
    "code": "Base.1.12.GeneralError",
    "message": "A general error has occurred. See ExtendedInfo for more information",
    "@Message.ExtendedInfo": [
      {
        "Message": "Unable to complete the operation because an invalid username and/or password is entered.",
        "MessageId": "IDRAC.2.9.LOG101",
        "Resolution": "Enter a valid username and password and retry the operation.",
        "Severity": "Critical"
      }
    ]
  }
}`

const bmcURL = "https://10.0.0.1/redfish/v1/"

func collectionError(failures map[string]error) error {
	collErr := schemas.NewCollectionError()
	for link, cause := range failures {
		collErr.Failures[link] = cause
	}
	return fmt.Errorf("reading systems: %w", collErr)
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "extended info wins over the general error",
			err:  schemas.ConstructError(400, []byte(dellGeneralError)),
			want: "Unable to complete the operation because the JobID entered is invalid.",
		},
		{
			name: "top level message when there is no extended info",
			err:  schemas.ConstructError(404, []byte(`{"error":{"code":"Base.1.0.ResourceMissing","message":"not found"}}`)),
			want: "not found",
		},
		{
			name: "http status text when the body says nothing",
			err:  schemas.ConstructError(500, []byte(`{}`)),
			want: "Internal Server Error",
		},
		{
			name: "http status text when the body is not a redfish error",
			err:  schemas.ConstructError(502, []byte(`<html><body>Bad Gateway from the web server</body></html>`)),
			want: "Bad Gateway",
		},
		{
			name: "status code when even that is unknown",
			err:  schemas.ConstructError(599, []byte(`{}`)),
			want: "bmc returned status 599",
		},
		{
			name: "a transport failure drops the method and url",
			err:  &url.Error{Op: "Get", URL: bmcURL, Err: &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}},
			want: "dial: connection refused",
		},
		{
			name: "a deadline",
			err:  &url.Error{Op: "Get", URL: bmcURL, Err: context.DeadlineExceeded},
			want: "context deadline exceeded",
		},
		{
			name: "an untrusted certificate",
			err:  &url.Error{Op: "Get", URL: bmcURL, Err: x509.UnknownAuthorityError{}},
			want: "x509: certificate signed by unknown authority",
		},
		{
			name: "our own errors pass through",
			err:  errors.New("bmc exposes no computer systems"),
			want: "bmc exposes no computer systems",
		},
		{
			name: "a collection failure reports its cause",
			err:  collectionError(map[string]error{"/redfish/v1/Systems/1": schemas.ConstructError(401, []byte(dellAuthError))}),
			want: "Unable to complete the operation because an invalid username and/or password is entered.",
		},
		{
			name: "several collection failures always report the same one",
			err: collectionError(map[string]error{
				"/redfish/v1/Systems/2": context.DeadlineExceeded,
				"/redfish/v1/Systems/1": errors.New("first"),
				"/redfish/v1/Systems/3": errors.New("last"),
			}),
			want: "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, errorMessage(tt.err))
		})
	}
}

// TestErrorMessageLocalGofishError covers gofish building an Error with no status code for a failure it detected itself, which is not an answer from the BMC
func TestErrorMessageLocalGofishError(t *testing.T) {
	assert.Contains(t, errorMessage(schemas.ConstructError(0, []byte("unable to execute request, no target provided"))), "no target provided")
}

// TestErrorMessageNeverLeaksRawBody is the reason a BMC's answer is not simply forwarded, since (*schemas.Error).Error() returns the whole response body
func TestErrorMessageNeverLeaksRawBody(t *testing.T) {
	err := schemas.ConstructError(400, []byte(dellGeneralError))

	assert.Contains(t, err.Error(), "@Message.ExtendedInfo", "the premise of this test")
	assert.NotContains(t, errorMessage(err), "@Message.ExtendedInfo")
}

func TestIsAuthFailure(t *testing.T) {
	rejected := schemas.ConstructError(401, []byte(dellAuthError))

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"401", rejected, true},
		{"403", schemas.ConstructError(403, []byte(`{}`)), true},
		{"500", schemas.ConstructError(500, []byte(`{}`)), false},
		{"wrapped", fmt.Errorf("connecting: %w", rejected), true},
		{"a collection member", collectionError(map[string]error{"/redfish/v1/Systems/1": rejected}), true},
		{"a transport failure", &url.Error{Op: "Get", URL: bmcURL, Err: syscall.ECONNREFUSED}, false},
		{"our own error", errors.New("no credentials"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAuthFailure(tt.err))
		})
	}
}
