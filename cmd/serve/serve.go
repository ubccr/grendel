package serve

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/model"
)

func init() {
	cmd.Root.AddCommand(serveCmd)
}

var (
	DB       *model.BuntStore
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Run services",
		Long:  `Run grendel services`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, cancel := NewInterruptContext()

			var wg sync.WaitGroup
			wg.Add(3)
			errs := make(chan error, 3)

			go func() {
				errs <- serveAPI(ctx)
				wg.Done()
			}()
			go func() {
				errs <- serveTFTP(ctx)
				wg.Done()
			}()
			go func() {
				errs <- serveDNS(ctx)
				wg.Done()
			}()

			// Fail if any servers error out
			err := <-errs

			cmd.Log.Infof("Waiting for all services to shutdown...")

			ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelShutdown()

			cancel()

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				cmd.Log.Info("All services shutdown")
			case <-ctxShutdown.Done():
				cmd.Log.Warning("Timeout reached")
			}

			return err
		},
	}
)

func NewInterruptContext() (context.Context, context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		cmd.Log.Debugf("Signal interrupt system call: %+v", oscall)
		cancel()
	}()

	return ctx, cancel
}

func init() {
	serveCmd.PersistentFlags().String("dbpath", ":memory:", "path to database file")
	viper.BindPFlag("dbpath", serveCmd.PersistentFlags().Lookup("dbpath"))

	// DNS
	serveCmd.PersistentFlags().String("dns-listen", "0.0.0.0:53", "address to listen on")
	serveCmd.PersistentFlags().Int("dns-ttl", 300, "ttl for dns records")
	viper.BindPFlag("dns.listen", serveCmd.PersistentFlags().Lookup("dns-listen"))
	viper.BindPFlag("dns.ttl", serveCmd.PersistentFlags().Lookup("dns-ttl"))

	// TFTP
	serveCmd.PersistentFlags().String("tftp-listen", "0.0.0.0:69", "address to listen on")
	viper.BindPFlag("tftp.listen", serveCmd.PersistentFlags().Lookup("tftp-listen"))

	// API
	serveCmd.PersistentFlags().String("api-listen", "0.0.0.0:6669", "address to listen on")
	viper.BindPFlag("api.listen", serveCmd.PersistentFlags().Lookup("api-listen"))
	serveCmd.PersistentFlags().String("api-socket", "", "path to unix socket")
	viper.BindPFlag("api.socket_path", serveCmd.PersistentFlags().Lookup("api-socket"))
	serveCmd.PersistentFlags().String("api-cert", "", "path to ssl cert")
	viper.BindPFlag("api.cert", serveCmd.PersistentFlags().Lookup("api-cert"))
	serveCmd.PersistentFlags().String("api-key", "", "path to ssl key")
	viper.BindPFlag("api.key", serveCmd.PersistentFlags().Lookup("api-key"))

	serveCmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		err := cmd.SetupLogging()
		if err != nil {
			return err
		}

		DB, err = model.NewBuntStore(viper.GetString("dbpath"))
		if err != nil {
			return err
		}

		cmd.Log.Infof("Using database path: %s", viper.GetString("dbpath"))

		return nil
	}

	serveCmd.PersistentPostRunE = func(command *cobra.Command, args []string) error {
		if DB != nil {
			cmd.Log.Info("Closing Database")
			err := DB.Close()
			if err != nil {
				return err
			}
		}

		return nil
	}
}
