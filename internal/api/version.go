// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"runtime"
	"runtime/debug"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/store/migrations"
	"github.com/ubccr/grendel/internal/version"
	"github.com/ubccr/grendel/pkg/model"
)

// GetVersion reports the running server's build. It touches neither the database nor any other service, so a 200 from it means the api server is up and nothing more
func (h *Handler) GetVersion(c fuego.ContextNoBody) (model.Version, error) {
	v := model.Version{
		Version:   version.Version,
		GoVersion: runtime.Version(),
		Schema:    migrations.SchemaVersion,
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				v.Commit = s.Value
			case "vcs.time":
				v.CommitTime = s.Value
			case "vcs.modified":
				v.Modified = s.Value == "true"
			}
		}
	}
	return v, nil
}
