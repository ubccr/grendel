package serve

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/frontend"
	"gopkg.in/tomb.v2"
)

func init() {
	frontendCmd.PersistentFlags().String("frontend-listen", "0.0.0.0:8080", "address to listen on")
	viper.BindPFlag("frontend.listen", frontendCmd.PersistentFlags().Lookup("frontend-listen"))
	frontendCmd.Flags().String("frontend-cert", "", "path to ssl cert")
	viper.BindPFlag("frontend.cert", frontendCmd.Flags().Lookup("frontend-cert"))
	frontendCmd.Flags().String("frontend-key", "", "path to ssl key")
	viper.BindPFlag("frontend.key", frontendCmd.Flags().Lookup("frontend-key"))

	serveCmd.AddCommand(frontendCmd)
}

// cli command: grendel serve frontend
var (
	frontendCmd = &cobra.Command{
		Use:   "frontend",
		Short: "Run Frontend WebUI",
		Long:  `Run Frontend WebUI`,
		RunE: func(command *cobra.Command, args []string) error {
			t := NewInterruptTomb()
			t.Go(func() error { return serveFrontend(t) })
			return t.Wait()
		},
	}
)

func serveFrontend(t *tomb.Tomb) error {
	frontendListen, err := GetListenAddress(viper.GetString("frontend.listen"))
	if err != nil {
		return err
	}

	frontendServer, err := frontend.NewServer(DB, frontendListen)
	if err != nil {
		return err
	}
	frontendServer.KeyFile = viper.GetString("frontend.key")
	frontendServer.CertFile = viper.GetString("frontend.cert")

	t.Go(func() error {
		time.Sleep(1 * time.Second)
		<-t.Dying()
		cmd.Log.Info("Shutting down frontend server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := frontendServer.Shutdown(ctxShutdown); err != nil {
			cmd.Log.Errorf("Failed shutting down frontend server: %s", err)
			return err
		}

		return nil
	})

	return frontendServer.Serve()
}
