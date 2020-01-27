package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dhcp"
)

func init() {
	pxeCmd.PersistentFlags().String("pxe-listen", "0.0.0.0:4011", "address to listen on")
	viper.BindPFlag("pxe.listen", pxeCmd.PersistentFlags().Lookup("pxe-listen"))

	serveCmd.AddCommand(pxeCmd)
}

var (
	pxeCmd = &cobra.Command{
		Use:   "pxe",
		Short: "Run DHCP PXE Boot server",
		Long:  `Run DHCP PXE Boot server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
			return servePXE(ctx)
		},
	}
)

func servePXE(ctx context.Context) error {
	srv, err := dhcp.NewPXEServer(DB, viper.GetString("pxe.listen"))
	if err != nil {
		return err
	}

	if err := srv.Serve(ctx); err != nil {
		return err
	}

	return nil
}
