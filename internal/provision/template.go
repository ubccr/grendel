// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha256_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/coreos/butane/config"
	"github.com/coreos/butane/config/common"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/pkg/model"
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
	"hasTag":                 hasTag,
	"Split":                  Split,
	"Join":                   Join,
	"Contains":               Contains,
	"ConfigValueStringSlice": ConfigValueStringSlice,
	"ConfigValueString":      ConfigValueString,
	"ConfigValueBool":        ConfigValueBool,
	"Add":                    Add,
	"Mul":                    Mul,
	"CryptSHA512":            CryptSHA512,
	"CryptSHA256":            CryptSHA256,
	"DellSHA256Password":     DellSHA256Password,
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

	ct := c.Response().Header().Get(echo.HeaderContentType)
	if ct == "" {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
	}

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

func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func Join(s []string, sep string) string {
	return strings.Join(s, sep)
}

func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func Add(a, b int) int {
	return a + b
}

func Mul(a, b int) int {
	return a * b
}

func ConfigValueString(key string) string {
	return viper.GetString(key)
}

func ConfigValueStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func ConfigValueBool(key string) bool {
	return viper.GetBool(key)
}

func CryptSHA512(pass, salt string) string {
	crypt := crypt.SHA512.New()
	hash512, _ := crypt.Generate([]byte(pass), []byte("$6$"+salt))
	return hash512
}

func CryptSHA256(pass, salt string) string {
	crypt := crypt.SHA256.New()
	hash256, _ := crypt.Generate([]byte(pass), []byte("$5$"+salt))
	return hash256
}

func DellSHA256Password(pass, salt string) string {
	h := sha256.New()
	payload := pass + salt
	h.Write([]byte(payload))
	return fmt.Sprintf("%x%x", h.Sum(nil), salt)
}
