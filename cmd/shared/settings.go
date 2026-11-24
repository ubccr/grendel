// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package shared

import (
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/settings"
)

var settingFlags = []*pflag.FlagSet{}

// BindSettingFlags gives fs a flag for every option of a scope that names one. A flag that is given overrides its option.
func BindSettingFlags(fs *pflag.FlagSet, scope settings.Scope) {
	for _, o := range settings.All() {
		if o.Flag == "" || o.Scope != scope {
			continue
		}
		switch o.Kind {
		case settings.KindBool:
			fs.Bool(o.Flag, o.Default.(bool), o.Summary)
		case settings.KindInt:
			fs.Int(o.Flag, o.Default.(int), o.Summary)
		case settings.KindStringList:
			fs.StringSlice(o.Flag, o.Default.([]string), o.Summary)
		default:
			fs.String(o.Flag, settings.Text(o.Default), o.Summary)
		}
		if o.Deprecated != "" {
			_ = fs.MarkDeprecated(o.Flag, o.Deprecated)
		}
	}
	settingFlags = append(settingFlags, fs)
}

// externalSetting finds an option in the layers that win over the database: a flag given on the command line, then the environment, then the config file
func externalSetting(o *settings.Option) (any, settings.Source, bool) {
	// --debug and --verbose are two flags for one setting, so they cannot be named by the option the way other flags are
	if o.Key == settings.LogLevel.Key {
		if debug {
			return "debug", settings.SourceFlag, true
		}
		if verbose {
			return "info", settings.SourceFlag, true
		}
	}
	if o.Flag != "" {
		for _, fs := range settingFlags {
			if f := fs.Lookup(o.Flag); f != nil && f.Changed {
				return f.Value.String(), settings.SourceFlag, true
			}
		}
	}
	if v, ok := os.LookupEnv(o.Env()); ok && v != "" {
		return v, settings.SourceEnv, true
	}
	if viper.InConfig(o.Key) {
		return viper.Get(o.Key), settings.SourceFile, true
	}
	return nil, "", false
}

// ReadSettingsFile reads a config file on its own, leaving the settings this process runs with untouched
func ReadSettingsFile(path string) (settings.File, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		return settings.File{}, err
	}
	return settings.File{
		Path: path,
		Keys: v.AllKeys(),
		Lookup: func(key string) (any, bool) {
			if !v.InConfig(key) {
				return nil, false
			}
			return v.Get(key), true
		},
	}, nil
}

// LoadSettings resolves every setting from the flags, the environment, the config file and, when b is not nil, the database
func LoadSettings(b settings.Backend) ([]settings.Problem, error) {
	settings.Configure(settings.Env{
		Lookup:     externalSetting,
		ConfigFile: viper.ConfigFileUsed(),
		Persistent: b != nil && !strings.Contains(settings.DBPath.Get(), ":memory:"),
	})
	return settings.Load(b)
}
