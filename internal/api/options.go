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
	"github.com/ubccr/grendel/pkg/model"
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

// redfishNulls holds the json name of every field in the redfish payloads whose zero value marshals as null
var redfishNulls = nilableFields(
	reflect.TypeOf(model.RedfishSystem{}),
	reflect.TypeOf(model.Event{}),
)

// nilableFields walks the given types and returns the json names of the fields that can marshal as null and not tagged omitempty.
func nilableFields(types ...reflect.Type) []string {
	var (
		names []string
		seen  = map[reflect.Type]bool{}
		walk  func(t reflect.Type)
	)

	walk = func(t reflect.Type) {
		for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || seen[t] {
			return
		}
		seen[t] = true

		for i := range t.NumField() {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			name, opts, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}

			switch field.Type.Kind() {
			case reflect.Slice, reflect.Map, reflect.Pointer, reflect.Interface:
				if !slices.Contains(strings.Split(opts, ","), "omitempty") {
					names = append(names, name)
				}
			}

			walk(field.Type)
		}
	}

	for _, t := range types {
		walk(t)
	}

	slices.Sort(names)
	return slices.Compact(names)
}

func schemaCustomizer() openapi3gen.SchemaCustomizerFn {
	return func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		if slices.Contains(redfishNulls, name) {
			schema.Nullable = true
		}

		// enum lists the values a string field takes, so the generated clients get a type for it rather than any string
		if enum := tag.Get("enum"); enum != "" {
			for _, v := range strings.Split(enum, ",") {
				schema.Enum = append(schema.Enum, v)
			}
		}

		st := tag.Get("oai3")
		for _, s := range strings.Split(st, ",") {
			switch s {
			case "nullable":
				schema.Nullable = true
			case "notnull":
				// a pointer field is nullable by default, which a required field meant to be set must not be
				schema.Nullable = false
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
