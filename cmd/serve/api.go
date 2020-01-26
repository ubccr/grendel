package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/api"
)

func init() {
	serveCmd.AddCommand(apiCmd)
}

var (
	apiCmd = &cobra.Command{
		Use:   "api",
		Short: "Run API server",
		Long:  `Run API server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
			return serveAPI(ctx)
		},
	}
)

func serveAPI(ctx context.Context) error {
	apiServer, err := api.NewServer(DB, viper.GetString("api.socket_path"), viper.GetString("api.listen"))
	if err != nil {
		return err
	}

	apiServer.KeyFile = viper.GetString("api.key")
	apiServer.CertFile = viper.GetString("api.cert")

	if err := apiServer.Serve(ctx); err != nil {
		return err
	}

	return nil
}
