package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			ctx, _ := NewInterruptContext()
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
