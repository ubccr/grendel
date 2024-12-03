// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"os"
	"path/filepath"

	"github.com/segmentio/ksuid"
)

type BootImageList []*BootImage

type BootImage struct {
	ID                 ksuid.KSUID       `json:"id"`
	Name               string            `json:"name" validate:"required"`
	KernelPath         string            `json:"kernel" validate:"required"`
	InitrdPaths        []string          `json:"initrd"`
	LiveImage          string            `json:"liveimg"`
	CommandLine        string            `json:"cmdline"`
	Verify             bool              `json:"verify"`
	ProvisionTemplate  string            `json:"provision_template"`
	ProvisionTemplates map[string]string `json:"provision_templates"`
	UserData           string            `json:"user_data"`
	Butane             string            `json:"butane"`
}

func NewBootImageList() BootImageList {
	return make(BootImageList, 0)
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

	if b.ProvisionTemplate != "" {
		if _, err := os.Stat(filepath.Join("/var/lib/grendel/templates", b.ProvisionTemplate)); err != nil {
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

	if b.UserData != "" {
		if _, err := os.Stat(filepath.Join("/var/lib/grendel/templates", b.UserData)); err != nil {
			return err
		}
	}

	if b.Butane != "" {
		if _, err := os.Stat(filepath.Join("/var/lib/grendel/templates", b.Butane)); err != nil {
			return err
		}
	}

	return nil
}
