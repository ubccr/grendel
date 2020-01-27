package host

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
)

var (
	hostCmd = &cobra.Command{
		Use:   "host",
		Short: "Host commands",
		Long:  `Host commands`,
	}
)

func init() {
	cmd.Root.AddCommand(hostCmd)
}
