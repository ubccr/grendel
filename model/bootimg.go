// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"os"

	"github.com/segmentio/ksuid"
)

type BootImageList []*BootImage

type BootImage struct {
	ID          ksuid.KSUID `json:"id"`
	Name        string      `json:"name"`
	KernelPath  string      `json:"kernel"`
	InitrdPaths []string    `json:"initrd"`
	LiveImage   string      `json:"liveimg"`
	InstallRepo string      `json:"install_repo"`
	CommandLine string      `json:"cmdline"`
	Verify      bool        `json:"verify"`
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

	return nil
}
