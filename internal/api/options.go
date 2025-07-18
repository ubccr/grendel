// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"reflect"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/go-fuego/fuego"
)

func setupOpenapiConfig(swaggerUI bool) fuego.OpenAPIConfig {
	return fuego.OpenAPIConfig{
		DisableSwaggerUI: !swaggerUI,
		JSONFilePath:     "./api/openapi.json",
		PrettyFormatJSON: true,
	}
}

func setupSecurity() openapi3.SecuritySchemes {
	return map[string]*openapi3.SecuritySchemeRef{
		"headerAuth": {
			Value: openapi3.NewSecurityScheme().
				WithType("http").
				WithIn("header").
				WithScheme("bearer").
				WithBearerFormat("JWT").
				WithDescription("API key header authentication."),
		},
		"cookieAuth": {
			Value: openapi3.NewSecurityScheme().
				WithType("http").
				WithIn("cookie").
				WithScheme("bearer").
				WithBearerFormat("JWT").
				WithDescription("API key cookie authentication."),
		},
	}
}

func schemaCustomizer() openapi3gen.SchemaCustomizerFn {
	return func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		redfishNulls := []string{"RelatedProperties", "HttpHeaders", "EnabledDaysOfMonth", "EnabledDaysOfWeek", "EnabledIntervals", "EnabledMonthsOfYear", "StepOrder"}
		if slices.Contains(redfishNulls, name) {
			schema.Nullable = true
		}

		st := tag.Get("oai3")
		for _, s := range strings.Split(st, ",") {
			switch s {
			case "nullable":
				schema.Nullable = true
			case "typeStrArr":
				schema.Items = openapi3.NewSchemaRef("", openapi3.NewStringSchema())
			case "typeStr":
				schema.Type = &openapi3.Types{"string"}
			case "formatNone":
				schema.Format = ""
			}
		}

		return nil
	}
}
