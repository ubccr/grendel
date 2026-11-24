// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package render

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha256_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/netbox"
)

var log = logger.GetLogger("TEMPLATE")

// NetBox is the client the NetBoxRenderConfig template function and the provision server share
var NetBox = netbox.NewClient()

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

// configValue reads a setting for a template. Secrets are refused, since a template's output is handed to whoever boots: a value a template needs belongs in provision.extra_vars.
func configValue(key string) (any, error) {
	o, ok := settings.Lookup(key)
	if !ok {
		return nil, fmt.Errorf("%s is not a setting", key)
	}
	if o.Secret {
		return nil, fmt.Errorf("%s is a secret and cannot be read from a template, pass it through provision.extra_vars", key)
	}
	v, _ := settings.Value(key)
	return v, nil
}

func ConfigValueString(key string) (string, error) {
	v, err := configValue(key)
	if err != nil {
		return "", err
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	return settings.Text(v), nil
}

func ConfigValueStringSlice(key string) ([]string, error) {
	v, err := configValue(key)
	if err != nil {
		return nil, err
	}
	list, ok := v.([]string)
	if !ok {
		return nil, fmt.Errorf("%s is not a list", key)
	}
	return list, nil
}

func ConfigValueBool(key string) (bool, error) {
	v, err := configValue(key)
	if err != nil {
		return false, err
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("%s is not true or false", key)
	}
	return b, nil
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
	res, err := NetBox.NetBoxFetchConfig(name)
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
