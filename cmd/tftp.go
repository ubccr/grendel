package cmd

import (
	"fmt"

	"github.com/ubccr/grendel/tftp"
	"github.com/urfave/cli"
)

func NewTFTPCommand() cli.Command {
	return cli.Command{
		Name:        "tftp",
		Usage:       "Start TFTP server",
		Description: "Start TFTP server",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "tftp-port",
				Value: 69,
				Usage: "tftp port to listen on",
			},
			cli.StringFlag{
				Name:  "listen-address",
				Value: "0.0.0.0",
				Usage: "address to listen on",
			},
		},
		Action: runTFTP,
	}
}

func runTFTP(c *cli.Context) error {
	listenAddress := c.String("listen-address")
	address := fmt.Sprintf("%s:%d", listenAddress, c.Int("tftp-port"))

	tftpServer, err := tftp.NewServer(address)
	if err != nil {
		return err
	}

	return tftpServer.Serve()
}
