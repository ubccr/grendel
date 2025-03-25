// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

//go:generate go run github.com/ogen-go/ogen/cmd/ogen@v1.10.0 --target ../../pkg/client --config ../../api/config.yaml --package client --clean ../../api/openapi.json

//go:generate yarn --cwd ../frontend run codegen
