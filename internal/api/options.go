// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"reflect"
	"slices"

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
		if name == "firmware" {
			schema.Type = &openapi3.Types{"string"}
		}
		if name == "ip" {
			schema.Type = &openapi3.Types{"string"}
		}
		if name == "mac" {
			schema.Format = ""
			schema.Type = &openapi3.Types{"string"}
		}
		if name == "uid" {
			schema.Type = &openapi3.Types{"string"}
			schema.Nullable = true
		}
		if name == "id" {
			schema.Nullable = true
		}
		if name == "provision_templates" {
			schema.Nullable = true
		}
		redfishNulls := []string{"RelatedProperties", "HttpHeaders", "EnabledDaysOfMonth", "EnabledDaysOfWeek", "EnabledIntervals", "EnabledMonthsOfYear", "StepOrder"}
		if slices.Contains(redfishNulls, name) {
			schema.Nullable = true
		}

		return nil
	}
}
