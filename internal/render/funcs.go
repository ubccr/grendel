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
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/logger"
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
