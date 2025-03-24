// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/ubccr/grendel/pkg/model"
)

func ParseRedfishError(err error) model.RedfishError {
	var e model.RedfishError
	if err == nil {
		return e
	}

	reg := regexp.MustCompile(`^[0-9]{3}`)
	e.Code = reg.FindString(err.Error())
	if e.Code != "" {
		test, _ := strings.CutPrefix(err.Error(), e.Code+":")

		json.Unmarshal([]byte(test), &e)
	}

	return e
}
