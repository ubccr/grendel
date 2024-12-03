// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"regexp"
	"strings"
)

type RedfishError struct {
	Code  string
	Error struct {
		MessageExtendedInfo []struct {
			Message                string
			MessageArgs            []string //?
			MessageArgsCount       int      `json:"MessageArgs.@odata.count"`
			MessageId              string
			RelatedProperties      []string //?
			RelatedPropertiesCount int      `json:"RelatedProperties.@odata.count"`
			Resolution             string
			Severity               string
		} `json:"@Message.ExtendedInfo"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func ParseRedfishError(err error) RedfishError {
	var e RedfishError
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
