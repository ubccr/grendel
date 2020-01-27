package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/provision"
)

func init() {
	provisionCmd.Flags().String("provision-listen", "0.0.0.0:80", "address to listen on")
	viper.BindPFlag("provision.listen", provisionCmd.Flags().Lookup("provision-listen"))
	provisionCmd.Flags().String("provision-cert", "", "path to ssl cert")
	viper.BindPFlag("provision.cert", provisionCmd.Flags().Lookup("provision-cert"))
	provisionCmd.Flags().String("provision-key", "", "path to ssl key")
	viper.BindPFlag("provision.key", provisionCmd.Flags().Lookup("provision-key"))
	provisionCmd.Flags().String("default-image", "", "default image name")
	viper.BindPFlag("provision.default_image", provisionCmd.Flags().Lookup("default-image"))
	provisionCmd.Flags().String("repo-dir", "", "path to repo dir")
	viper.BindPFlag("provision.repo_dir", provisionCmd.Flags().Lookup("repo-dir"))

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
	srv.RepoDir = viper.GetString("provision.repo_dir")

	if err := srv.Serve(ctx, viper.GetString("provision.default_image")); err != nil {
		return err
	}

	return nil
}
