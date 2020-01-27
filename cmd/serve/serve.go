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
			wg.Add(6)
			errs := make(chan error, 6)

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
			go func() {
				errs <- serveProvision(ctx)
				wg.Done()
			}()
			go func() {
				errs <- serveDHCP(ctx)
				wg.Done()
			}()
			go func() {
				errs <- servePXE(ctx)
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
