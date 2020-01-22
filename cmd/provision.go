package cmd

import (
	"github.com/ubccr/grendel/provision"
	"github.com/urfave/cli/v2"
)

func NewProvisionCommand() *cli.Command {
	return &cli.Command{
		Name:        "provision",
		Usage:       "Start HTTP provision server",
		Description: "Start HTTP provision server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "http-port",
				Value: 80,
				Usage: "http port to listen on",
			},
			&cli.StringFlag{
				Name:  "http-scheme",
				Value: "http",
				Usage: "http scheme",
			},
			&cli.StringFlag{
				Name:  "listen-address",
				Value: "0.0.0.0",
				Usage: "IPv4 address to listen on",
			},
			&cli.StringFlag{
				Name:  "cert",
				Usage: "Path to certificate",
			},
			&cli.StringFlag{
				Name:  "key",
				Usage: "Path to private key",
			},
			&cli.StringFlag{
				Name:  "repo-dir",
				Usage: "Path to repo dir",
			},
		},
		Action: runProvision,
	}
}

func runProvision(c *cli.Context) error {
	httpPort := c.Int("http-port")
	if c.IsSet("cert") && c.IsSet("key") && !c.IsSet("http-port") {
		httpPort = 443
	}

	provisionServer, err := provision.NewServer(DB, c.String("listen-address"), httpPort)
	if err != nil {
		return err
	}

	provisionServer.KeyFile = c.String("key")
	provisionServer.CertFile = c.String("cert")
	provisionServer.RepoDir = c.String("repo-dir")

	return provisionServer.Serve()
}
