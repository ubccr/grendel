package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/provision"
)

func init() {
	provisionCmd.PersistentFlags().String("provision-listen", "0.0.0.0:80", "address to listen on")
	viper.BindPFlag("provision.listen", provisionCmd.PersistentFlags().Lookup("provision-listen"))
	provisionCmd.PersistentFlags().String("provision-cert", "", "path to ssl cert")
	viper.BindPFlag("provision.cert", provisionCmd.PersistentFlags().Lookup("provision-cert"))
	provisionCmd.PersistentFlags().String("provision-key", "", "path to ssl key")
	viper.BindPFlag("provision.key", provisionCmd.PersistentFlags().Lookup("provision-key"))

	serveCmd.AddCommand(provisionCmd)
}

var (
	provisionCmd = &cobra.Command{
		Use:   "provision",
		Short: "Run Provision server",
		Long:  `Run Provision server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
			return serveProvision(ctx)
		},
	}
)

func serveProvision(ctx context.Context) error {
	srv, err := provision.NewServer(DB, viper.GetString("provision.listen"))
	if err != nil {
		return err
	}

	srv.KeyFile = viper.GetString("provision.key")
	srv.CertFile = viper.GetString("provision.cert")

	if err := srv.Serve(ctx); err != nil {
		return err
	}

	return nil
}
