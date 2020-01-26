package serve

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/dns"
)

func init() {
	serveCmd.AddCommand(dnsCmd)
}

var (
	dnsCmd = &cobra.Command{
		Use:   "dns",
		Short: "Run DNS server",
		Long:  `Run DNS server`,
		RunE: func(command *cobra.Command, args []string) error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)

			ctx, cancel := context.WithCancel(context.Background())

			go func() {
				oscall := <-c
				cmd.Log.Debugf("Signal interrupt system call: %+v", oscall)
				cancel()
			}()

			return serveDNS(ctx)
		},
	}
)

func serveDNS(ctx context.Context) error {
	dnsServer, err := dns.NewServer(DB, viper.GetString("dns.listen"), viper.GetInt("dns.ttl"))
	if err != nil {
		return err
	}

	if err := dnsServer.Serve(ctx); err != nil {
		return err
	}

	return nil
}
