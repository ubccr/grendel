package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/tftp"
)

func init() {
	tftpCmd.PersistentFlags().String("tftp-listen", "0.0.0.0:69", "address to listen on")
	viper.BindPFlag("tftp.listen", tftpCmd.PersistentFlags().Lookup("tftp-listen"))

	serveCmd.AddCommand(tftpCmd)
}

var (
	tftpCmd = &cobra.Command{
		Use:   "tftp",
		Short: "Run TFTP server",
		Long:  `Run TFTP server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
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