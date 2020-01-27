package host

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

var (
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show hosts",
		Long:  `Show hosts`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := client.NewClient()
			if err != nil {
				return err
			}

			var hostList model.HostList

			if len(args) == 1 && strings.ToLower(args[0]) == "all" {
				hostList, err = gc.Hosts()
				if err != nil {
					return err
				}
			} else {
				ns, err := nodeset.NewNodeSet(strings.Join(args, ","))
				if err != nil {
					return err
				}

				if ns.Len() == 0 {
					return errors.New("Node nodes in nodeset")
				}

				hostList, err = gc.FindHosts(ns)
				if err != nil {
					return err
				}
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
