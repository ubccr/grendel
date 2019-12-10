package cmd

import (
	"github.com/ubccr/grendel/api"
	"github.com/urfave/cli"
)

func NewAPICommand() cli.Command {
	return cli.Command{
		Name:        "api",
		Usage:       "Start API HTTP server",
		Description: "Start API HTTP server",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "http-port",
				Value: 80,
				Usage: "http port to listen on",
			},
			cli.StringFlag{
				Name:  "http-scheme",
				Value: "http",
				Usage: "http scheme",
			},
			cli.StringFlag{
				Name:  "listen-address",
				Value: "0.0.0.0",
				Usage: "IPv4 address to listen on",
			},
			cli.StringFlag{
				Name:  "cert",
				Usage: "Path to certificate",
			},
			cli.StringFlag{
				Name:  "key",
				Usage: "Path to private key",
			},
		},
		Action: runAPI,
	}
}

func runAPI(c *cli.Context) error {
	if c.IsSet("cert") && c.IsSet("key") && !c.IsSet("http-port") {
		c.Set("http-port", "443")
	}

	apiServer, err := api.NewServer(DB, c.String("listen-address"), c.Int("http-port"))
	if err != nil {
		return err
	}

	apiServer.KeyFile = c.String("key")
	apiServer.CertFile = c.String("cert")

	return apiServer.Serve()
}
