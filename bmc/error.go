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

func ParseRedfishError(err error) (RedfishError, error) {
	var e RedfishError
	if err == nil {
		return e, nil
	}

	reg := regexp.MustCompile(`^[0-9]{3}`)
	e.Code = reg.FindString(err.Error())
	test, _ := strings.CutPrefix(err.Error(), e.Code+":")

	if err := json.Unmarshal([]byte(test), &e); err != nil {
		return e, err
	}

	return e, nil
}
