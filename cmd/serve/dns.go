package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dns"
)

func init() {
	dnsCmd.PersistentFlags().String("dns-listen", "0.0.0.0:53", "address to listen on")
	dnsCmd.PersistentFlags().Int("dns-ttl", 300, "ttl for dns records")
	viper.BindPFlag("dns.listen", dnsCmd.PersistentFlags().Lookup("dns-listen"))
	viper.BindPFlag("dns.ttl", dnsCmd.PersistentFlags().Lookup("dns-ttl"))

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
