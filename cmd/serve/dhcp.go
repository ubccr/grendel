package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dhcp"
)

func init() {
	dhcpCmd.PersistentFlags().String("dhcp-listen", "0.0.0.0:67", "address to listen on")
	viper.BindPFlag("dhcp.listen", dhcpCmd.PersistentFlags().Lookup("dhcp-listen"))

	serveCmd.AddCommand(dhcpCmd)
}

var (
	dhcpCmd = &cobra.Command{
		Use:   "dhcp",
		Short: "Run DHCP server",
		Long:  `Run DHCP server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
			return serveDHCP(ctx)
		},
	}
)

func serveDHCP(ctx context.Context) error {
	srv, err := dhcp.NewServer(DB, viper.GetString("dhcp.listen"))
	if err != nil {
		return err
	}

	if err := srv.Serve(ctx); err != nil {
		return err
	}

	return nil
}
