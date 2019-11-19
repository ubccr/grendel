package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/cmd"
	"github.com/urfave/cli"
)

var release = "(version not set)"

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
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.WarnLevel)
		}

		return nil
	}
	app.Commands = []cli.Command{
		cmd.NewCertsCommand(),
		cmd.NewServeCommand(),
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
