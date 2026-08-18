// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/version"
	"github.com/ubccr/grendel/pkg/model"
)

// GetVersion reports the running server's build version. It touches neither the database nor any other service, so a 200 from it means the api server is up and nothing more
func (h *Handler) GetVersion(c fuego.ContextNoBody) (model.Version, error) {
	return model.Version{Version: version.Version}, nil
}
