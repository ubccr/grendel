package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/logger"
	"github.com/urfave/cli"
)

var release = "(version not set)"
var log = logger.GetLogger("main")

func main() {
	app := cli.NewApp()
	app.Name = "grendel"
	app.Version = release
	app.Usage = "provisioning system for high-performance Linux clusters"
	app.Author = "Andrew E. Bruno"
	app.Email = "aebruno2@buffalo.edu"
	app.Flags = []cli.Flag{
		&cli.BoolFlag{Name: "verbose", Usage: "Print verbose messages"},
	}
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("verbose") {
			log.Logger.SetLevel(logrus.DebugLevel)
		} else {
			log.Logger.SetLevel(logrus.WarnLevel)
		}

		return nil
	}
	app.Commands = []cli.Command{
		cmd.NewCertsCommand(),
		cmd.NewServeCommand(),
	}
	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
