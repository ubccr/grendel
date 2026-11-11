// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"

	"github.com/stmcginnis/gofish/schemas"
)

// errorMessage parses and attempts to return the most relavent error
func errorMessage(err error) string {
	if err == nil {
		return ""
	}

	err = parseCollectionErrors(err)

	// parse redfish errors
	if err, ok := errors.AsType[*schemas.Error](err); ok && err.HTTPReturnedStatusCode != 0 {
		return parseRedfishError(err)
	}

	// simplify HTTP connection errors
	if err, ok := errors.AsType[*url.Error](err); ok {
		return err.Err.Error()
	}

	return err.Error()
}

// parseCollectionErrors returns the first sorted error if the request was to a collection
func parseCollectionErrors(err error) error {
	if err, ok := errors.AsType[*schemas.CollectionError](err); ok && len(err.Failures) > 0 {
		return err.Failures[slices.Sorted(maps.Keys(err.Failures))[0]]
	}

	return err
}

// parseRedfishError tries to return the error with the most relavent information
func parseRedfishError(e *schemas.Error) string {
	if len(e.ExtendedInfos) > 0 && e.ExtendedInfos[0].Message != "" {
		return e.ExtendedInfos[0].Message
	}

	if e.Message != "" && e.Code != "" {
		return e.Message
	}

	if text := http.StatusText(e.HTTPReturnedStatusCode); text != "" {
		return text
	}

	return fmt.Sprintf("bmc returned status %d", e.HTTPReturnedStatusCode)
}
