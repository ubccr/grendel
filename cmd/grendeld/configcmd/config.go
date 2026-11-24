// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package configcmd

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendeld/serve"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/internal/settings"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Move settings between a config file and the database",
		Long:  "Move settings between a config file and the database. These commands open the database directly, so they work whether or not grendeld is running. Where grendeld runs as its own user, run them as root, since they read its config file and write its database. Use `grendel config` to view and change individual settings.",
		PersistentPreRunE: func(command *cobra.Command, args []string) error {
			// the root command only stops printing usage once logging is set up, which is after this check
			command.SilenceUsage, command.SilenceErrors = true, true
			if err := checkAccess(settings.List().ConfigFile, settings.DBPath.Get()); err != nil {
				return err
			}
			problems, err := serve.OpenDB()
			if err != nil {
				return err
			}
			if !settings.List().Persistent {
				return errors.New("the database is in memory, so there is nothing to keep settings in: set dbpath first")
			}
			for _, p := range problems {
				shared.Log.Warn(p)
			}
			return nil
		},
		PersistentPostRun: func(command *cobra.Command, args []string) {
			serve.CloseDB()
		},
	}

	importApply     bool
	importEntries   []string
	importRemaining string
	importReplace   bool
	importCmd       = &cobra.Command{
		Use:   "import [file]",
		Short: "Import settings from a config file into the database",
		Long: `Import settings from a config file into the database. The file is the one grendeld runs with unless another is given.

Without --apply this only shows what would happen to each thing the file sets: import or skip. With --apply, every setting marked import is stored in the database.

Secrets and settings that can only be set in the config file are always skipped, as are values already at their default. A value under a deprecated key is imported under the key that replaced it.

Entries of provision.extra_vars are skipped unless named with --entry, since it often holds passwords.

--replace rewrites the file to hold only what was skipped and still has to be set in a file, keeping the original beside it with .bak added to its name. --remaining writes that to another file instead and leaves the original alone. What is kept is: secrets, file-only settings, client settings and the provision.extra_vars entries that were not imported. Defaults, deprecated keys and unknown keys are left out. Restart grendeld afterwards, as a value in the config file wins over the database.`,
		Example: `  sudo grendeld config import
  sudo grendeld config import --apply --replace
  sudo grendeld config import old.toml --apply --remaining /etc/grendel/grendel.toml
  sudo grendeld config import --apply --entry provision.extra_vars=timezone`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			path := settings.List().ConfigFile
			if len(args) == 1 {
				path = args[0]
			}
			if path == "" {
				return errors.New("no config file is in use, so name the file to import")
			}
			file, err := shared.ReadSettingsFile(path)
			if err != nil {
				return asRoot(err)
			}

			entries := map[string][]string{}
			for _, e := range importEntries {
				key, name, ok := strings.Cut(e, "=")
				if !ok {
					return fmt.Errorf("--entry %q is not setting=entry", e)
				}
				entries[key] = append(entries[key], name)
			}

			if importReplace && !importApply {
				return errors.New("--replace needs --apply, as it removes from the file what the import stores")
			}
			// checked before anything is stored, so a file that cannot be written does not leave the import half done
			switch {
			case importReplace:
				err = canReplace(path)
			case importRemaining != "":
				err = canCreate(importRemaining)
			}
			if err != nil {
				return fmt.Errorf("nothing was changed: %w", err)
			}

			plan, err := settings.PlanImport(file, entries)
			if err != nil {
				return err
			}

			t := shared.NewTable([]shared.Column{{Name: "Key"}, {Name: "Value"}, {Name: "Action"}, {Name: "Description"}}, nil)
			for _, item := range plan.Items {
				action := "skip"
				if item.Action == settings.ImportReady {
					action = "import"
				}
				t.AppendRow(item.Key, item.Value, action, item.Note)
			}
			t.Print()
			fmt.Println()

			if invalid := plan.Invalid(); len(invalid) > 0 && (importApply || importRemaining != "") {
				keys := make([]string, len(invalid))
				for i, item := range invalid {
					keys[i] = item.Key
				}
				return fmt.Errorf("nothing was changed: fix or remove the invalid values in %s first: %s", path, strings.Join(keys, ", "))
			}

			if !importApply {
				fmt.Printf("%d setting(s) will be imported from %s into %s. Run again with --apply to import them.\n", plan.Count(), path, absolute(settings.DBPath.Get()))
			} else {
				refused, err := plan.Apply(serve.DB, "grendeld config import")
				if err != nil {
					return err
				}
				if len(refused) > 0 {
					for _, key := range slices.Sorted(maps.Keys(refused)) {
						fmt.Fprintf(os.Stderr, "%s: %s\n", key, refused[key])
					}
					return errors.New("nothing was imported")
				}
				fmt.Printf("Imported %d setting(s) from %s into %s.\n", plan.Count(), path, absolute(settings.DBPath.Get()))
			}

			switch {
			case importReplace:
				return replaceFile(plan, path)
			case importRemaining != "":
				return writeRemaining(plan, path)
			case importApply && plan.Count() > 0:
				fmt.Println("The config file wins over the database, so remove the imported lines from it and restart grendeld. --replace does that for you.")
			}
			return nil
		},
	}

	exportEffective bool
	exportCmd       = &cobra.Command{
		Use:   "export",
		Short: "Print settings as a config file, without secrets",
		Long:  "Print the settings the database stores as a config file. With --effective, print every value in use, wherever it comes from. Secrets are never included.",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			fmt.Print(settings.Export(exportEffective))
			return nil
		},
	}
)

const remainingHeader = "# Settings that have to be set in this file. The rest are in the database: see `grendel config list`.\n\n"

func asRoot(err error) error {
	if errors.Is(err, fs.ErrPermission) && os.Geteuid() != 0 {
		return fmt.Errorf("%w: run this as root, for example with sudo", err)
	}
	return err
}

func absolute(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

func checkAccess(configFile, dbPath string) error {
	// a config file that is found but cannot be read leaves every setting at its default, dbpath among them
	if configFile != "" {
		f, err := os.Open(configFile)
		if err != nil {
			return asRoot(err)
		}
		f.Close()
	}

	if strings.Contains(dbPath, ":memory:") {
		return nil
	}
	// opening a database that is not there creates an empty one, owned by whoever runs this rather than by grendeld
	f, err := os.OpenFile(dbPath, os.O_RDWR, 0)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("there is no database at %s: start grendeld once to create it, or check dbpath", absolute(dbPath))
	}
	if err != nil {
		return asRoot(err)
	}
	return f.Close()
}

func canCreate(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%s is already there, so move it away first", path)
	}
	dir := filepath.Dir(path)
	probe, err := os.CreateTemp(dir, filepath.Base(path)+".check-*")
	if err != nil {
		return asRoot(fmt.Errorf("cannot write in %s: %w", dir, cause(err)))
	}
	probe.Close()
	return os.Remove(probe.Name())
}

// cause drops the path from a file error, which here names a probe file rather than anything the user asked for
func cause(err error) error {
	var path *fs.PathError
	var link *os.LinkError
	switch {
	case errors.As(err, &path):
		return path.Err
	case errors.As(err, &link):
		return link.Err
	}
	return err
}

// canReplace tries what replaceFile does beside source, so a step added there needs its check here
func canReplace(source string) error {
	for _, next := range []string{source + ".bak", source + ".new"} {
		if err := canCreate(next); err != nil {
			return err
		}
	}
	// the backup is a hard link, which the kernel refuses for a file the user neither owns nor can write
	probe := source + ".check-link"
	if err := os.Link(source, probe); err != nil {
		return asRoot(fmt.Errorf("cannot keep %s as a backup: %w", source, cause(err)))
	}
	return os.Remove(probe)
}

// create writes a file that holds secrets, and never replaces an existing one
func create(path, content string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func writeRemaining(plan *settings.ImportPlan, source string) error {
	remaining := plan.Remaining()
	if remaining == "" {
		fmt.Printf("Nothing in %s has to stay in a config file, so %s was not written.\n", source, importRemaining)
		return nil
	}
	if err := create(importRemaining, remainingHeader+remaining); err != nil {
		return err
	}
	fmt.Printf("Wrote %s with the settings that have to stay in a config file. Replace %s with it and restart grendeld.\n", importRemaining, source)
	return nil
}

func replaceFile(plan *settings.ImportPlan, source string) error {
	backup, next := source+".bak", source+".new"
	if err := create(next, remainingHeader+plan.Remaining()); err != nil {
		return err
	}
	// grendeld often runs as its own user and reads the file through its group, which a file created here as root would not allow
	if info, err := os.Stat(source); err == nil {
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			_ = os.Chown(next, int(st.Uid), int(st.Gid))
		}
		_ = os.Chmod(next, info.Mode().Perm())
	}
	// the original keeps its name until the new file is complete, so a failure part way leaves a working config
	if err := os.Link(source, backup); err != nil {
		os.Remove(next)
		return err
	}
	if err := os.Rename(next, source); err != nil {
		return err
	}
	fmt.Printf("Rewrote %s with only the settings that have to stay in a config file, and kept the original as %s. Restart grendeld to use the database's values.\n", source, backup)
	return nil
}

func init() {
	importCmd.Flags().BoolVar(&importApply, "apply", false, "import the settings rather than only listing them")
	importCmd.Flags().StringArrayVar(&importEntries, "entry", nil, "import one entry of provision.extra_vars, as provision.extra_vars=<entry>")
	importCmd.Flags().BoolVar(&importReplace, "replace", false, "rewrite the file to hold only what could not be imported, keeping the original as <file>.bak")
	importCmd.Flags().StringVar(&importRemaining, "remaining", "", "write a config file holding only what could not be imported")
	importCmd.MarkFlagsMutuallyExclusive("replace", "remaining")
	exportCmd.Flags().BoolVar(&exportEffective, "effective", false, "every value in use rather than only what the database stores")

	configCmd.AddCommand(importCmd, exportCmd)
	shared.Root.AddCommand(configCmd)
}
