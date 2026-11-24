// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Server settings stored in the database",
	}

	listSection string
	listSource  string
	listPending bool
	listChanged bool
	listJSON    bool
	listTable   *shared.TableFlags
	listColumns = []shared.Column{
		{Name: "Key"},
		{Name: "Value"},
		{Name: "Source"},
		{Name: "State"},
	}
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List settings, their values and where each value comes from",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			res, err := apiclient.API.GETV1Settings(command.Context(), client.GETV1SettingsParams{})
			if err != nil {
				return apiclient.NewApiError(err)
			}

			var shown []client.ListingSectionsItemOptionsItem
			for _, section := range res.Sections {
				if listSection != "" && !strings.EqualFold(section.ID.Value, listSection) {
					continue
				}
				for _, o := range section.Options {
					if listSource != "" && o.Source.Value != listSource {
						continue
					}
					if listPending && !o.Pending.Value {
						continue
					}
					if listChanged && !isSet(o) {
						continue
					}
					shown = append(shown, o)
				}
			}
			if listJSON {
				return shared.OutputJSON(shown)
			}

			t := listTable.Table()
			for _, o := range shown {
				t.AppendRow(o.Key.Value, displayValue(o), o.Source.Value, state(o))
			}
			t.Print()

			warnNotPersistent(res)
			return nil
		},
	}

	describeCmd = &cobra.Command{
		Use:   "describe <key>",
		Short: "Explain a setting: what it does, what it accepts and how it is set",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			res, err := apiclient.API.GETV1Settings(command.Context(), client.GETV1SettingsParams{})
			if err != nil {
				return apiclient.NewApiError(err)
			}
			for _, o := range options(res) {
				if o.Key.Value != args[0] {
					continue
				}
				describe(o, res.ConfigFile.Value)
				return nil
			}
			return fmt.Errorf("unknown setting %s", args[0])
		},
	}

	getCmd = &cobra.Command{
		Use:   "get <key>",
		Short: "Print the value of a setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			res, err := apiclient.API.GETV1Settings(command.Context(), client.GETV1SettingsParams{})
			if err != nil {
				return apiclient.NewApiError(err)
			}
			for _, o := range options(res) {
				if o.Key.Value != args[0] {
					continue
				}
				if o.Secret.Value {
					return fmt.Errorf("%s is a secret and is never shown", args[0])
				}
				fmt.Println(apiclient.Text(o.Value))
				return nil
			}
			return fmt.Errorf("unknown setting %s", args[0])
		},
	}

	setCmd = &cobra.Command{
		Use:   "set <key> <value> | <key>=<value>...",
		Short: "Store settings in the database",
		Long:  "Store settings in the database. A list is written with commas or as JSON, and a map as JSON. Nothing is saved unless every value is valid.",
		Example: `  grendel config set dns.forward 1.1.1.1:53
  grendel config set dns.ttl=600 events.retention=720h
  grendel config set dhcp.dns_servers 10.0.0.2,10.0.0.3
  grendel config set dhcp.subnets.rack-f.gateway=10.17.41.254/23 dhcp.subnets.rack-f.mtu=9000
  grendel config set loggers.tftp off`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			set := client.PatchSettingsRequestSet{}
			// each setting is either key=value in one argument, or a key followed by its value
			for i := 0; i < len(args); i++ {
				key, value, ok := strings.Cut(args[i], "=")
				if !ok {
					if i+1 == len(args) {
						return fmt.Errorf("%s has no value", args[i])
					}
					value = args[i+1]
					i++
				}
				// sent as text, which the server reads the way it reads a flag or the environment
				set[key], _ = json.Marshal(value)
			}

			res, err := apiclient.API.PATCHV1Settings(command.Context(), &client.PatchSettingsRequest{
				Set: client.NewOptPatchSettingsRequestSet(set),
			}, client.PATCHV1SettingsParams{})
			if err != nil {
				return refusal(err)
			}

			for _, o := range options(res) {
				if _, changed := set[o.Key.Value]; !changed {
					continue
				}
				delete(set, o.Key.Value)
				if o.Pending.Value {
					fmt.Printf("%s saved, restart grendeld to apply (running with %s)\n", o.Key.Value, apiclient.Text(o.Running))
				} else {
					fmt.Printf("%s saved, applies now\n", o.Key.Value)
				}
			}
			// what is left named an entry of a setting rather than a setting, such as dhcp.subnets.rack-f.mtu
			for _, key := range slices.Sorted(maps.Keys(set)) {
				fmt.Printf("%s saved, applies now\n", key)
			}
			warnNotPersistent(res)
			return nil
		},
	}

	unsetCmd = &cobra.Command{
		Use:   "unset <key>...",
		Short: "Remove settings from the database, returning them to their default",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			res, err := apiclient.API.PATCHV1Settings(command.Context(), &client.PatchSettingsRequest{
				Unset: args,
			}, client.PATCHV1SettingsParams{})
			if err != nil {
				return refusal(err)
			}

			known := map[string]client.ListingSectionsItemOptionsItem{}
			for _, o := range options(res) {
				known[o.Key.Value] = o
			}
			for _, key := range args {
				if o, ok := known[key]; ok {
					fmt.Printf("%s removed, now %s (%s)\n", key, orEmpty(displayValue(o)), o.Source.Value)
				} else {
					fmt.Printf("%s removed\n", key)
				}
			}
			return nil
		},
	}
)

func init() {
	listTable = shared.RegisterTableFlags(listCmd, listColumns)
	listCmd.Flags().StringVar(&listSection, "section", "", "only this section, such as dhcp or bmc")
	listCmd.Flags().StringVar(&listSource, "source", "", "only settings whose value comes from: default, database, file, env or flag")
	listCmd.Flags().BoolVar(&listPending, "pending", false, "only settings waiting for a restart")
	listCmd.Flags().BoolVar(&listChanged, "changed", false, "only settings that are not at their default")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "print everything known about each setting as JSON")

	configCmd.AddCommand(listCmd, getCmd, describeCmd, setCmd, unsetCmd)
	shared.Root.AddCommand(configCmd)
}

func options(res *client.Listing) []client.ListingSectionsItemOptionsItem {
	var out []client.ListingSectionsItemOptionsItem
	for _, section := range res.Sections {
		out = append(out, section.Options...)
	}
	return out
}

func isSet(o client.ListingSectionsItemOptionsItem) bool {
	return o.Source.Value != "default" && o.Source.Value != "inherited"
}

// writable reports whether the database may hold the setting, which it may not for a secret or one only read from the config file
func writable(o client.ListingSectionsItemOptionsItem) bool {
	return !o.Secret.Value && !o.External.Value
}

func displayValue(o client.ListingSectionsItemOptionsItem) string {
	if o.Secret.Value {
		if isSet(o) {
			return "(set)"
		}
		return "(not set)"
	}
	return apiclient.Text(o.Value)
}

func state(o client.ListingSectionsItemOptionsItem) string {
	var states []string
	if !writable(o) {
		states = append(states, "set outside the database")
	}
	if o.Pending.Value {
		states = append(states, fmt.Sprintf("restart pending, running with %s", apiclient.Text(o.Running)))
	}
	if o.Shadowed.Value {
		states = append(states, fmt.Sprintf("database value %s is overridden", apiclient.Text(o.Stored)))
	}
	if len(o.FileKeys) > 0 {
		states = append(states, "from the config file: "+strings.Join(o.FileKeys, ", "))
	}
	if o.Source.Value == "inherited" {
		states = append(states, "follows "+o.InheritFrom.Value)
	}
	if o.Deprecated.Value != "" {
		states = append(states, "deprecated, "+o.Deprecated.Value)
	}
	return strings.Join(states, "; ")
}

func describe(o client.ListingSectionsItemOptionsItem, configFile string) {
	fmt.Printf("%s\n%s\n\n%s\n\n", o.Key.Value, o.Summary.Value, o.Help.Value)

	kind := o.Kind.Value
	if len(o.Enum) > 0 {
		kind = "one of " + strings.Join(o.Enum, ", ")
	}
	if o.Unit.Value != "" {
		kind += ", in " + o.Unit.Value
	}
	fmt.Printf("  %-10s %s\n", "Accepts", kind)
	for _, f := range o.Fields {
		fmt.Printf("  %-10s   %s (%s) %s\n", "", f.Name.Value, f.Kind.Value, f.Help.Value)
	}
	if !o.Secret.Value {
		fmt.Printf("  %-10s %s\n", "Default", orEmpty(apiclient.Text(o.Default)))
		fmt.Printf("  %-10s %s (from the %s)\n", "Value", orEmpty(apiclient.Text(o.Value)), o.Source.Value)
	} else {
		fmt.Printf("  %-10s %s, never shown\n", "Value", displayValue(o))
	}
	if o.Apply.Value == "restart" {
		fmt.Printf("  %-10s after grendeld restarts\n", "Applies")
	} else {
		fmt.Printf("  %-10s as soon as it is saved\n", "Applies")
	}

	var ways []string
	if writable(o) {
		ways = append(ways, "grendel config set "+o.Key.Value)
	}
	if o.Flag.Value != "" {
		ways = append(ways, "--"+o.Flag.Value)
	}
	if o.Env.Value != "" {
		ways = append(ways, o.Env.Value)
	}
	if configFile == "" {
		configFile = "grendel.toml"
	}
	ways = append(ways, configFile)
	fmt.Printf("  %-10s %s\n", "Set with", strings.Join(ways, ", "))
	if state := state(o); state != "" {
		fmt.Printf("  %-10s %s\n", "State", state)
	}
}

func orEmpty(s string) string {
	if s == "" {
		return "(empty)"
	}
	return s
}

func warnNotPersistent(res *client.Listing) {
	if !res.Persistent.Value {
		fmt.Println("warning: grendeld is using an in-memory database, saved settings are lost when it stops")
	}
}

// refusal turns the per-setting reasons of a rejected change into one error
func refusal(apiError error) error {
	var t *client.HTTPErrorStatusCode
	if !errors.As(apiError, &t) || len(t.GetResponse().Errors.Value) == 0 {
		return apiclient.NewApiError(apiError)
	}

	reasons := make([]string, 0, len(t.GetResponse().Errors.Value))
	for _, item := range t.GetResponse().Errors.Value {
		reasons = append(reasons, fmt.Sprintf("  %s: %s", item.Name.Value, item.Reason.Value))
	}
	return fmt.Errorf("nothing was saved:\n%s", strings.Join(reasons, "\n"))
}
