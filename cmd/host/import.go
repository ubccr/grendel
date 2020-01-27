package host

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/cmd"
)

var (
	importCmd = &cobra.Command{
		Use:   "import",
		Short: "import hosts",
		Long:  `import hosts`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := client.NewClient()
			if err != nil {
				return err
			}

			for _, name := range args {
				file, err := os.Open(name)
				if err != nil {
				}
				defer file.Close()

				cmd.Log.Infof("Processing file: %s", name)
				err = gc.StoreHostsReader(file)
				if err != nil {
					return err
				}

				cmd.Log.Infof("Successfully imported hosts from: %s", name)
			}
			return nil

		},
	}
)

func init() {
	hostCmd.AddCommand(importCmd)
}
