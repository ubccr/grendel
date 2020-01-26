package serve

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/tftp"
)

func init() {
	serveCmd.AddCommand(tftpCmd)
}

var (
	tftpCmd = &cobra.Command{
		Use:   "tftp",
		Short: "Run TFTP server",
		Long:  `Run TFTP server`,
		RunE: func(command *cobra.Command, args []string) error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)

			ctx, cancel := context.WithCancel(context.Background())

			go func() {
				oscall := <-c
				cmd.Log.Debugf("Signal interrupt system call: %+v", oscall)
				cancel()
			}()

			return serveTFTP(ctx)
		},
	}
)

func serveTFTP(ctx context.Context) error {
	tftpServer, err := tftp.NewServer(viper.GetString("tftp.listen"))
	if err != nil {
		return err
	}

	if err := tftpServer.Serve(ctx); err != nil {
		return err
	}

	return nil
}
