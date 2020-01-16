package main

import (
	"fmt"
	"io/ioutil"
	golog "log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/util"
	"github.com/urfave/cli/v2"
)

var release = "(version not set)"
var log = logger.GetLogger("main")

func init() {
	viper.SetConfigName("grendel")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/grendel/")
}

func main() {
	app := cli.NewApp()
	app.Name = "grendel"
	app.Version = release
	app.Usage = "provisioning system for high-performance Linux clusters"
	app.Authors = []*cli.Author{{Name: "Andrew E. Bruno", Email: "aebruno2@buffalo.edu"}}
	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "conf,c", Usage: "Path to conf file"},
		&cli.BoolFlag{Name: "verbose", Usage: "Print verbose messages"},
		&cli.BoolFlag{Name: "debug", Usage: "Print debug messages"},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.Logger.SetLevel(logrus.DebugLevel)
		} else if c.Bool("verbose") {
			log.Logger.SetLevel(logrus.InfoLevel)
		} else {
			log.Logger.SetLevel(logrus.WarnLevel)
		}
		golog.SetOutput(ioutil.Discard)

		conf := c.String("conf")
		if len(conf) > 0 {
			viper.SetConfigFile(conf)

			err := viper.ReadInConfig()
			if err != nil {
				return fmt.Errorf("Failed reading config file: %s", err)
			}
		}

		if !viper.IsSet("provision.secret") {
			secret, err := util.GenerateSecret(32)
			if err != nil {
				return err
			}

			viper.Set("provision.secret", secret)
		}

		if !viper.IsSet("api.secret") {
			secret, err := util.GenerateSecret(32)
			if err != nil {
				return err
			}

			viper.Set("api.secret", secret)
		}

		return nil
	}
	app.Commands = []*cli.Command{
		cmd.NewCertsCommand(),
		cmd.NewServeCommand(),
		cmd.NewHostCommand(),
	}
	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
