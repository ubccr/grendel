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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/logger"
)

func OutputJSON(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	return enc.Encode(data)
}

func SetupLogging() error {
	logger.Apply()
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

	err := viper.ReadInConfig()
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		// a missing config file is worth mentioning to someone running a command, but not to a shell sourcing the completion script, which puts it in front of them every time they open a terminal
		if !completing() {
			Log.Warn(err)
		}
	} else if err != nil {
		Log.Errorln(err)
	}

	// only the client's own settings are checked here, as grendeld checks the rest once its database is open
	problems, err := LoadSettings(nil)
	if err != nil {
		Log.Fatal(err)
	}
	for _, p := range problems {
		if strings.HasPrefix(p.Key, "client.") {
			Log.Warn(p)
		}
	}
}
