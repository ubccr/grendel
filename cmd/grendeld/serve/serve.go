// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	DB            store.Store
	hostsFile     string
	imagesFile    string
	listenAddress string
	serveCmd      = &cobra.Command{
		Use:   "serve [service]...",
		Short: "Run grendel services",
		Args:  cobra.ArbitraryArgs,
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			names := make([]string, 0, len(registry))
			for _, s := range services() {
				names = append(names, s.Name)
			}

			return names, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(command *cobra.Command, args []string) error {
			// cobra skips PersistentPostRunE when RunE returns an error, so the close belongs here where a failed start still reaches it
			defer CloseDB()

			names := args
			if len(names) == 0 {
				names = settings.Services.Get()
			}

			svcs, err := selectServices(names)
			if err != nil {
				return err
			}

			if imagesFile != "" {
				err := loadImageJSON()
				if err != nil {
					return err
				}
			}
			if hostsFile != "" {
				err := loadHostJSON()
				if err != nil {
					return err
				}
			}

			return run(command.Context(), svcs)
		},
	}
)

func init() {
	serveCmd.PersistentFlags().StringVar(&hostsFile, "hosts", "", "path to hosts file")
	serveCmd.PersistentFlags().StringVar(&imagesFile, "images", "", "path to boot images file")
	serveCmd.PersistentFlags().StringVar(&listenAddress, "listen", "", "listen address")

	serveCmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		problems, err := OpenDB()
		if err != nil {
			return err
		}
		for _, p := range problems {
			// a bad value in the database is skipped so it can be fixed from the running server, but one in the config file, the environment or a flag stops the start, as nothing would be using what was asked for
			if p.Source != settings.SourceDatabase {
				return fmt.Errorf("invalid setting: %s", p)
			}
			shared.Log.Warn(p)
		}
		logger.Apply()
		settings.OnChange(logger.Apply)
		importHint()
		return nil
	}

	shared.BindSettingFlags(serveCmd.PersistentFlags(), settings.ScopeServer)
	shared.Root.AddCommand(serveCmd)
}

// OpenDB opens the database into DB and loads the settings it holds, returning the values that were skipped as invalid
func OpenDB() ([]settings.Problem, error) {
	if err := shared.SetupLogging(); err != nil {
		return nil, err
	}
	if dbType := settings.DBType.Get(); dbType != "sqlite" {
		return nil, fmt.Errorf("unsupported dbtype: %s", dbType)
	}
	db, err := sqlstore.New(settings.DBPath.Get())
	if err != nil {
		return nil, err
	}
	DB = db
	shared.Log.Infof("Using database path: %s", settings.DBPath.Get())

	problems, err := shared.LoadSettings(DB)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	return problems, nil
}

func importHint() {
	path := settings.List().ConfigFile
	if path == "" {
		return
	}
	file, err := shared.ReadSettingsFile(path)
	if err != nil {
		return
	}
	if plan, err := settings.PlanImport(file, nil); err == nil && plan.Count() > 0 {
		shared.Log.Warnf("%d setting(s) in %s can be kept in the database instead: run `grendeld config import` to see them", plan.Count(), path)
	}
}

// CloseDB is deferred by RunE rather than run from a post run hook, so that a service that fails to start still checkpoints and releases the database
func CloseDB() {
	if DB == nil {
		return
	}

	shared.Log.Info("Closing Database")

	if err := DB.Close(); err != nil {
		shared.Log.Errorf("Failed closing database: %s", err)
	}
}

func loadHostJSON() error {
	jsonBlob, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	var hostList model.HostList
	err = json.Unmarshal(jsonBlob, &hostList)
	if err != nil {
		return err
	}

	err = DB.StoreHosts(hostList)
	if err != nil {
		return err
	}

	shared.Log.Infof("Successfully loaded %d hosts", len(hostList))
	return nil
}

func loadImageJSON() error {
	jsonBlob, err := os.ReadFile(imagesFile)
	if err != nil {
		return err
	}

	var imageList model.BootImageList
	err = json.Unmarshal(jsonBlob, &imageList)
	if err != nil {
		return err
	}

	err = DB.StoreBootImages(imageList)
	if err != nil {
		return err
	}

	shared.Log.Infof("Successfully loaded %d boot images", len(imageList))
	return nil
}

// getListenAddress returns the listen address a setting holds, with its host swapped for the one given with --listen unless the service's own listen flag was given
func getListenAddress(key string) string {
	v, _ := settings.Value(key)
	address, _ := v.(string)
	if listenAddress == "" {
		return address
	}

	if serveCmd.Flags().Changed(strings.Replace(key, ".", "-", 1)) {
		return address
	}

	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}

	return fmt.Sprintf("%s:%s", listenAddress, port)
}
