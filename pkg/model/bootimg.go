// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/segmentio/ksuid"
)

type BootImageList []*BootImage

type BootImage struct {
	ID                 int64             `json:"id" oai3:"nullable"`
	UID                ksuid.KSUID       `json:"uid" oai3:"typeStr,nullable"`
	Name               string            `json:"name" validate:"required"`
	Version            string            `json:"version"`
	KernelPath         string            `json:"kernel" validate:"required"`
	InitrdPaths        []string          `json:"initrd"`
	LiveImage          string            `json:"liveimg"` // Deprecated: not stored since sql migration
	CommandLine        string            `json:"cmdline"`
	Verify             bool              `json:"verify"`
	ProvisionTemplates map[string]string `json:"provision_templates" oai3:"nullable"`
}

func NewBootImageList() BootImageList {
	return make(BootImageList, 0)
}

func (b *BootImage) Scan(value interface{}) error {
	data, ok := value.(string)
	if !ok {
		return errors.New("incompatible type")
	}
	var image BootImage
	err := json.Unmarshal([]byte(data), &image)
	if err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}

	*b = image
	return nil
}
