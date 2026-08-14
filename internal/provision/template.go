// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha256_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/coreos/butane/config"
	"github.com/coreos/butane/config/common"
	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/pkg/model"
)

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
	"indent":                 indent,
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
	"NetBoxRenderConfig":     NetBoxRenderConfig,
}

type TemplateRenderer struct {
	templates atomic.Pointer[template.Template]
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
	t := &TemplateRenderer{
		templates: atomic.Pointer[template.Template]{},
	}
	return t, t.reload()
}

func buildTemplates() (*template.Template, error) {
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

	if viper.IsSet("provision.templates_dir") {
		dir := filepath.Join(viper.GetString("provision.templates_dir"), "*.tmpl")
		matches, err := filepath.Glob(dir)
		if err != nil {
			return nil, err
		}

		if len(matches) == 0 {
			log.WithField("dir", dir).Warn("no additional templates found in provision.templates_dir")
		} else {
			tmpl, err = tmpl.Funcs(funcMap).ParseGlob(dir)
			if err != nil {
				return nil, err
			}
		}
	}

	log.Debug(tmpl.DefinedTemplates())
	return tmpl, nil
}

func (t *TemplateRenderer) reload() error {
	tmpl, err := buildTemplates()
	if err != nil {
		return err
	}

	t.templates.Store(tmpl)
	return nil
}

func (t *TemplateRenderer) Watch(ctx context.Context) error {
	if !viper.IsSet("provision.templates_dir") {
		return nil
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = w.Add(viper.GetString("provision.templates_dir"))
	if err != nil {
		return fmt.Errorf("failed to monitor template directory: %w", err)
	}

	go func() {
		defer w.Close()

		var debounce *time.Timer
		var debounceChan <-chan time.Time

		for {
			select {
			case <-ctx.Done():
				return

			case event, ok := <-w.Events:
				if !ok {
					return
				}

				if filepath.Ext(event.Name) != ".tmpl" {
					continue
				}
				if !event.Has(fsnotify.Write | fsnotify.Create | fsnotify.Remove | fsnotify.Rename) {
					continue
				}

				if debounce == nil {
					debounce = time.NewTimer(300 * time.Millisecond)
					debounceChan = debounce.C
					continue
				}

				debounce.Reset(300 * time.Millisecond)

			case <-debounceChan:
				err := t.reload()
				if err != nil {
					log.WithError(err).Error("template hot reload failed")
					return
				}

				log.Info("Successfully hot reloaded templates")

			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.WithError(err).Warn("failed to watch template directory")
			}
		}
	}()

	return nil
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	if viewContext, isMap := data.(map[string]any); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	ct := c.Response().Header().Get(echo.HeaderContentType)
	if ct == "" {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
	}

	test := t.templates.Load()
	if test != nil {
		return test.ExecuteTemplate(w, name, data)
	}
	return errors.New("failed to load templates")
}

func (t *TemplateRenderer) RenderIgnition(code int, name string, data any, c echo.Context) error {
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

func NetBoxRenderConfig(name string) string {
	res, err := netBoxClient.NetBoxFetchConfig(name)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	text, err := io.ReadAll(res.Body)
	if err != nil {
		log.WithFields(logrus.Fields{
			"name": name,
			"err":  err,
		}).Error("failed to read response body from netbox render-config")
		return ""
	}

	return string(text)
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.ReplaceAll(v, "\n", "\n"+pad)
}
