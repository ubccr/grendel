// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/segmentio/ksuid"
)

type BootImageList []*BootImage

type BootImage struct {
	ID                 int64             `json:"id"`
	UID                ksuid.KSUID       `json:"uid"`
	Name               string            `json:"name" validate:"required"`
	KernelPath         string            `json:"kernel" validate:"required"`
	InitrdPaths        []string          `json:"initrd"`
	LiveImage          string            `json:"liveimg"`
	CommandLine        string            `json:"cmdline"`
	Verify             bool              `json:"verify"`
	ProvisionTemplates map[string]string `json:"provision_templates"`
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

func (b *BootImage) CheckPathsExist() error {
	if _, err := os.Stat(b.KernelPath); err != nil {
		return err
	}

	for _, i := range b.InitrdPaths {
		if _, err := os.Stat(i); err != nil {
			return err
		}
	}

	if b.LiveImage != "" {
		if _, err := os.Stat(b.LiveImage); err != nil {
			return err
		}
	}

	if b.ProvisionTemplates != nil {
		for _, tmpl := range b.ProvisionTemplates {
			if _, err := os.Stat(filepath.Join("/var/lib/grendel/templates", tmpl)); err != nil {
				return err
			}
		}
	}

	return nil
}
