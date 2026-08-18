// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package shared

import (
	"encoding/json"
	"errors"
	"io"
	golog "log"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/config"
	"github.com/ubccr/grendel/internal/util"
)

func OutputJSON(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	return enc.Encode(data)
}

func SetupLogging() error {
	if debug {
		Log.Logger.SetLevel(logrus.DebugLevel)
	} else if verbose {
		Log.Logger.SetLevel(logrus.InfoLevel)
	} else {
		Log.Logger.SetLevel(logrus.WarnLevel)
	}
	golog.SetOutput(io.Discard)

	Log.Infof("Using config file: %s", viper.ConfigFileUsed())

	Root.SilenceUsage = true
	Root.SilenceErrors = true

	return nil
}

// completing reports whether cobra's completion machinery started this process, either a shell sourcing the completion script or a tab press. Both are meant to be silent: the script's output is fed straight to the shell and a tab press throws stderr away
func completing() bool {
	if len(os.Args) < 2 {
		return false
	}

	switch os.Args[1] {
	case "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return true
	}

	return false
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			Log.Fatal(err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			Log.Fatal(err)
		}

		viper.AddConfigPath("/etc/grendel/")
		viper.AddConfigPath(home)
		viper.AddConfigPath(cwd)
		viper.SetConfigName("grendel.toml")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("grendel")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		// a missing config file is worth mentioning to someone running a command, but not to a shell sourcing the completion script, which puts it in front of them every time they open a terminal
		if !completing() {
			Log.Warn(err)
		}
	} else if err != nil {
		Log.Errorln(err)
	}

	if !viper.IsSet("api.secret") {
		secret, err := util.GenerateSecret(32)
		if err != nil {
			Log.Fatal(err)
		}

		viper.Set("api.secret", secret)
	}

	err = config.ParseConfigs()
	if err != nil {
		Log.Fatal(err)
	}
}
