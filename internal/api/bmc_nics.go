// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/pkg/model"
)

func (h *Handler) BmcNICs(c fuego.ContextNoBody) (model.BmcResponse[model.BmcNICList], error) {
	return runBmc(h, c, "bmc nics", nil,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[model.BmcNICList] {
			return h.Redfish.NICs(ctx, t)
		},
		func(data *model.BmcNICList) *model.BmcNICList { return data },
	)
}
