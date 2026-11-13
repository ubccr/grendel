// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package fwcmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/internal/firmware"
)

var (
	exportForce bool
	listDir     string
	listFilter  []string
	listOptions *shared.TableOptions
	listColumns = []shared.Column{
		{Name: "Build"},
		{Name: "Source"},
		{Name: "Size"},
		{Name: "SHA-256"},
		{Name: "Signed"},
		{Name: "Current"},
	}
	firmwareCmd = &cobra.Command{
		Use:   "firmware",
		Short: "iPXE firmware commands",
		Long:  "Write out the iPXE builds embedded in grendeld, so they can be signed for UEFI secure boot and served from tftp.firmware_dir in place of the embedded ones",
	}
	exportCmd = &cobra.Command{
		Use:   "export <dir>",
		Short: "Write the embedded EFI builds to a directory",
		Long: `Write the embedded EFI builds of iPXE to a directory, named as Grendel serves them, such as snponly-x86_64.efi. The BIOS builds are left out, since secure boot does not apply to them.

Sign the files in place, point tftp.firmware_dir at the directory, and restart grendeld. A build missing from the directory is served from the embedded copy.`,
		Example: `  grendeld firmware export /var/lib/grendel/firmware
  sbsign --key db.key --cert db.crt --output /var/lib/grendel/firmware/snponly-x86_64.efi /var/lib/grendel/firmware/snponly-x86_64.efi`,
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			paths, err := firmware.Export(args[0], exportForce)
			for _, p := range paths {
				fmt.Println(p)
			}
			return err
		},
	}
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "Show where each iPXE build is served from, and whether it is signed",
		Long:  "Show where each iPXE build is served from, and whether it is signed. Current says whether a build in the directory is this grendeld's own, signatures aside, so a copy signed before an upgrade shows as no. The directory is tftp.firmware_dir unless --dir is given, so this checks a directory before grendeld is restarted to serve it",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			dir := listDir
			if !command.Flags().Changed("dir") {
				dir = viper.GetString("tftp.firmware_dir")
			}

			set, err := firmware.Load(dir)
			if err != nil {
				return err
			}

			t := shared.NewTable(listColumns, listFilter).Options(*listOptions).Empty("-")
			for _, b := range firmware.Builds() {
				f, ok := set[b]
				if !ok {
					t.AppendRow(b.String(), "missing", "", "", "", "")
					continue
				}
				sum := sha256.Sum256(f.Data)
				t.AppendRow(b.String(), source(b, f), strconv.Itoa(len(f.Data)), hex.EncodeToString(sum[:]), signedState(b, f), currentState(f))
			}
			t.Print()

			for _, path := range firmware.Ignored(dir) {
				fmt.Fprintf(os.Stderr, "Ignoring %s, BIOS builds are always served from the embedded copy\n", path)
			}
			return nil
		},
	}
)

// source is where a build comes from. A BIOS build says it is embedded only, since the directory is never read for one
func source(b firmware.Build, f firmware.File) string {
	switch {
	case f.Path != "":
		return f.Path
	case b.IsEFI():
		return "embedded"
	default:
		return "embedded only"
	}
}

// signedState only colours the EFI builds, since a BIOS build being unsigned is not something to fix
func signedState(b firmware.Build, f firmware.File) string {
	switch {
	case !b.IsEFI():
		return "n/a"
	case f.Signed:
		return text.FgHiGreen.Sprint("yes")
	default:
		return text.FgHiYellow.Sprint("no")
	}
}

func init() {
	exportCmd.Flags().BoolVar(&exportForce, "force", false, "overwrite files already in the directory, which may be signed copies")
	firmwareCmd.AddCommand(exportCmd)

	listOptions = shared.RegisterTableFlags(listCmd)
	listCmd.Flags().StringVar(&listDir, "dir", "", "directory to check instead of tftp.firmware_dir")
	listCmd.Flags().StringSliceVar(&listFilter, "filter", nil, "filter table columns by name")
	listCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(listColumns))
	firmwareCmd.AddCommand(listCmd)

	shared.Root.AddCommand(firmwareCmd)
}

// currentState says whether a build from the directory is this grendeld's, signatures aside. It is blank for an embedded build, and for a file this grendeld has no build of its own to compare with
func currentState(f firmware.File) string {
	switch f.Version {
	case firmware.VersionCurrent:
		return text.FgHiGreen.Sprint("yes")
	case firmware.VersionOutdated:
		return text.FgHiYellow.Sprint("no")
	default:
		return ""
	}
}
