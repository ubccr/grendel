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
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
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
			defer closeDB()

			names := args
			if len(names) == 0 {
				names = viper.GetStringSlice("services")
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
	serveCmd.PersistentFlags().String("dbtype", "sqlite", "database backend to use")
	viper.BindPFlag("dbtype", serveCmd.PersistentFlags().Lookup("dbtype"))
	serveCmd.PersistentFlags().String("dbpath", ":memory:", "path to database file")
	viper.BindPFlag("dbpath", serveCmd.PersistentFlags().Lookup("dbpath"))
	serveCmd.PersistentFlags().StringVar(&hostsFile, "hosts", "", "path to hosts file")
	serveCmd.PersistentFlags().StringVar(&imagesFile, "images", "", "path to boot images file")
	serveCmd.PersistentFlags().StringSlice("services", []string{}, "enabled services, ignored when services are named on the command line")
	serveCmd.PersistentFlags().StringVar(&listenAddress, "listen", "", "listen address")
	viper.BindPFlag("services", serveCmd.PersistentFlags().Lookup("services"))
	serveCmd.PersistentFlags().Duration("shutdown-timeout", defaultShutdownTimeout, "how long to wait for a service to stop")
	viper.BindPFlag("shutdown_timeout", serveCmd.PersistentFlags().Lookup("shutdown-timeout"))

	serveCmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		err := cmd.SetupLogging()
		if err != nil {
			return err
		}

		dbType := viper.GetString("dbtype")

		switch dbType {
		case "sqlite":
			DB, err = sqlstore.New(viper.GetString("dbpath"))
			if err != nil {
				return err
			}
		default:
			cmd.Log.Fatalf("unsupported dbtype: %s", dbType)
		}

		cmd.Log.Infof("Using %s database path: %s", dbType, viper.GetString("dbpath"))
		return nil
	}

	cmd.Root.AddCommand(serveCmd)
}

// closeDB is deferred by RunE rather than run from a post run hook, so that a service that fails to start still checkpoints and releases the database
func closeDB() {
	if DB == nil {
		return
	}

	cmd.Log.Info("Closing Database")

	if err := DB.Close(); err != nil {
		cmd.Log.Errorf("Failed closing database: %s", err)
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

	cmd.Log.Infof("Successfully loaded %d hosts", len(hostList))
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

	for _, i := range imageList {
		err = i.CheckPathsExist()
		if err != nil {
			return err
		}
	}

	err = DB.StoreBootImages(imageList)
	if err != nil {
		return err
	}

	cmd.Log.Infof("Successfully loaded %d boot images", len(imageList))
	return nil
}

// GetListenAddress takes a viper key and returns a listen address
//
// Local `--tftp-listen` flags override the global `--listen` flag
func getListenAddress(key string) string {
	address := viper.GetString(key)
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
