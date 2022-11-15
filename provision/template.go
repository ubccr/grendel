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

package provision

import (
	"bytes"
	_ "embed"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/coreos/butane/config"
	"github.com/coreos/butane/config/common"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/model"
)

const defaultTemplateGlob = "/var/lib/grendel/templates/*.tmpl"

//go:embed templates/ipxe.tmpl
var ipxeTmpl string

//go:embed templates/kickstart.tmpl
var kickstartTmpl string

//go:embed templates/user-data.tmpl
var userDataTmpl string

//go:embed templates/meta-data.tmpl
var metaDataTmpl string

//go:embed templates/butane.tmpl
var butaneTmpl string

// Template functions
var funcMap = template.FuncMap{
	"hasTag":            hasTag,
	"Split":             split,
	"ConfigValueString": ConfigValueString,
	"ConfigValueBool":   ConfigValueBool,
}

type TemplateRenderer struct {
	templates *template.Template
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
	tmpl, err := template.New("ipxe.tmpl").Funcs(funcMap).Parse(ipxeTmpl)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("kickstart.tmpl").Funcs(funcMap).Parse(kickstartTmpl)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("user-data.tmpl").Funcs(funcMap).Parse(userDataTmpl)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("meta-data.tmpl").Funcs(funcMap).Parse(metaDataTmpl)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("butane.tmpl").Funcs(funcMap).Parse(butaneTmpl)
	if err != nil {
		return nil, err
	}

	matches, err := filepath.Glob(defaultTemplateGlob)
	if err != nil {
		return nil, err
	}

	if len(matches) > 0 {
		tmpl, err = tmpl.Funcs(funcMap).ParseGlob(defaultTemplateGlob)
		if err != nil {
			return nil, err
		}
	}

	t := &TemplateRenderer{
		templates: tmpl,
	}

	return t, nil
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)

	return t.templates.ExecuteTemplate(w, name, data)
}

func (t *TemplateRenderer) RenderIgnition(code int, name string, data interface{}, c echo.Context) error {
	buf := new(bytes.Buffer)
	err := t.Render(buf, name, data, c)
	if err != nil {
		return err
	}

	options := common.TranslateBytesOptions{
		Pretty: false,
	}

	// TODO: how should we handle warnings in the translation?
	dataOut, _, err := config.TranslateBytes(buf.Bytes(), options)
	if err != nil {
		return err
	}

	return c.HTMLBlob(code, dataOut)
}

func hasTag(host model.Host, tag string) bool {
	return host.HasTags(tag)
}

func split(s, sep string) []string {
	return strings.Split(s, sep)
}

func ConfigValueString(key string) string {
	return viper.GetString(key)
}

func ConfigValueBool(key string) bool {
	return viper.GetBool(key)
}
