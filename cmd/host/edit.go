package host

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/util"
)

var (
	editCmd = &cobra.Command{
		Use:   "edit",
		Short: "edit hosts",
		Long:  `edit hosts`,
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

			hostList, err := gc.FindHosts(ns)
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(hostList, "", "    ")
			if err != nil {
				return err
			}

			newData, err := util.CaptureInputFromEditor(data)
			if err != nil {
				return err
			}

			var check model.HostList
			err = json.Unmarshal(newData, &check)
			if err != nil {
				return fmt.Errorf("Invalid JSON. Not saving changes: %w", err)
			}

			err = gc.StoreHostsReader(bytes.NewReader(newData))
			if err != nil {
				return err
			}

			fmt.Printf("Successfully saved %d hosts\n", len(check))

			return nil
		},
	}
)

func init() {
	hostCmd.AddCommand(editCmd)
}
