package cmd

import (
	"github.com/ubccr/grendel/api"
	"github.com/urfave/cli/v2"
)

func NewAPICommand() *cli.Command {
	return &cli.Command{
		Name:        "api",
		Usage:       "Start API HTTP server",
		Description: "Start API HTTP server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "api-port",
				Value: 6667,
				Usage: "http port to listen on",
			},
			&cli.StringFlag{
				Name:  "api-scheme",
				Value: "http",
				Usage: "api http scheme",
			},
			&cli.StringFlag{
				Name:  "api-address",
				Value: "0.0.0.0",
				Usage: "IPv4 address to listen on",
			},
			&cli.StringFlag{
				Name:  "socket-path",
				Usage: "Unix domain socket path",
			},
			&cli.StringFlag{
				Name:  "cert",
				Usage: "Path to certificate",
			},
			&cli.StringFlag{
				Name:  "key",
				Usage: "Path to private key",
			},
		},
		Action: runAPI,
	}
}

func runAPI(c *cli.Context) error {
	httpPort := c.Int("api-port")

	apiServer, err := api.NewServer(DB, c.String("socket-path"), c.String("api-address"), httpPort)
	if err != nil {
		return err
	}

	apiServer.KeyFile = c.String("key")
	apiServer.CertFile = c.String("cert")

	return apiServer.Serve()
}
