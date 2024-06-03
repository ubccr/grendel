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

package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/model"
	"gopkg.in/tomb.v2"
)

var (
	DB            model.DataStore
	hostsFile     string
	imagesFile    string
	listenAddress string
	serveCmd      = &cobra.Command{
		Use:   "serve",
		Short: "Run services",
		Long:  `Run grendel services`,
		RunE: func(command *cobra.Command, args []string) error {
			if hostsFile != "" {
				err := loadHostJSON()
				if err != nil {
					return err
				}
			}
			if imagesFile != "" {
				err := loadImageJSON()
				if err != nil {
					return err
				}
			}

			return runServices()
		},
	}
)

func init() {
	serveCmd.PersistentFlags().String("dbpath", ":memory:", "path to database file")
	viper.BindPFlag("dbpath", serveCmd.PersistentFlags().Lookup("dbpath"))
	serveCmd.PersistentFlags().String("dbtype", "buntdb", "database type (buntdb, rqlite)")
	viper.BindPFlag("dbtype", serveCmd.PersistentFlags().Lookup("dbtype"))
	serveCmd.PersistentFlags().String("dbaddr", "http://localhost:4001", "rqlite address")
	viper.BindPFlag("dbaddr", serveCmd.PersistentFlags().Lookup("dbaddr"))
	serveCmd.PersistentFlags().StringVar(&hostsFile, "hosts", "", "path to hosts file")
	serveCmd.PersistentFlags().StringVar(&imagesFile, "images", "", "path to boot images file")
	serveCmd.PersistentFlags().StringSlice("services", []string{}, "enabled services")
	serveCmd.PersistentFlags().StringVar(&listenAddress, "listen", "", "listen address")
	viper.BindPFlag("services", serveCmd.PersistentFlags().Lookup("services"))

	serveCmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		err := cmd.SetupLogging()
		if err != nil {
			return err
		}

		DB, err = model.NewDataStore(viper.GetString("dbtype"), viper.GetString("dbpath"), viper.GetString("dbaddr"))
		if err != nil {
			return err
		}

		cmd.Log.Infof("Using database: %s, file_path: %s, addr: %s", viper.GetString("dbtype"), viper.GetString("dbpath"), viper.GetString("dbaddr"))

		return nil
	}

	serveCmd.PersistentPostRunE = func(command *cobra.Command, args []string) error {
		if DB != nil {
			cmd.Log.Info("Closing Database")
			err := DB.Close()
			if err != nil {
				return err
			}
		}

		return nil
	}

	cmd.Root.AddCommand(serveCmd)
}

func loadHostJSON() error {
	file, err := os.Open(hostsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonBlob, err := ioutil.ReadAll(file)
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
	file, err := os.Open(imagesFile)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonBlob, err := ioutil.ReadAll(file)
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

func runServices() error {
	t := NewInterruptTomb()
	t.Go(func() error {
		t.Go(func() error { return serveTFTP(t) })
		t.Go(func() error { return serveDNS(t) })
		t.Go(func() error { return serveDHCP(t) })
		t.Go(func() error { return servePXE(t) })
		t.Go(func() error { return serveAPI(t) })
		t.Go(func() error { return serveProvision(t) })
		t.Go(func() error { return serveFrontend(t) })
		return nil
	})
	return t.Wait()
}

func NewInterruptTomb() *tomb.Tomb {
	t := &tomb.Tomb{}
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		select {
		case <-t.Dying():
		case <-sigint:
			cmd.Log.Debug("Caught interrupt signal")
			t.Kill(nil)
		}
	}()

	return t
}

func GetListenAddress(address string) (string, error) {
	if listenAddress == "" {
		return address, nil
	}

	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", listenAddress, port), nil
}

func NewInterruptContext() (context.Context, context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		cmd.Log.Debugf("Signal interrupt system call: %+v", oscall)
		cancel()
	}()

	return ctx, cancel
}
