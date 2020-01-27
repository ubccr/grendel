package host

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/nodeset"
)

var (
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show hosts",
		Long:  `Show hosts`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			ns, err := nodeset.NewNodeSet(strings.Join(args, ","))
			if err != nil {
				return err
			}

			if ns.Len() == 0 {
				return errors.New("Node nodes in nodeset")
			}

			gc, err := client.NewClient()
			if err != nil {
				return err
			}

			hostList, err := gc.HostFind(ns)
			if err != nil {
				return err
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "    ")
			if err := enc.Encode(hostList); err != nil {
				return err
			}

			return nil

		},
	}
)

func init() {
	hostCmd.AddCommand(showCmd)
}
