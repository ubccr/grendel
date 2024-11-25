// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package bmc

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/bmc"
	"golang.org/x/net/html/charset"
	"golang.org/x/sync/errgroup"
)

var (
	firmwareCmd = &cobra.Command{
		Use:   "firmware",
		Short: "BMC Firmware commands",
		Long:  `BMC Firmware commands`,
	}

	firmwareCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Check for updates on Dell servers",
		Long:  `Check for updates on Dell servers`,
		RunE: func(command *cobra.Command, args []string) error {
			return runFirmwareCheck()
		},
	}
	firmwareCheckShort            bool
	firmwareCheckCatalog          string
	firmwareCheckCatalogDownload  bool
	firmwareCheckFirmwareDownload string

	firmwareUpgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade firmware packages",
		Long:  "Upgrade firmware packages via Redfish Simple Update. BMC upgrades MUST be done separately from other updates as it will cancel any other scheduled or in progress upgrades. It is recommended to clear any old jobs from the BMC with the 'grendel bmc job clear' command first. Upgrades that require a reboot will show as a 'scheduled' job which will require a reboot to start the upgrade, the 'grendel bmc power cycle' command can be used to reboot the host.",
		RunE: func(command *cobra.Command, args []string) error {
			return runFirmwareUpgrade()
		},
	}
	firmwareUpgradePackages []string
	firmwareUpgradePath     string
)

func init() {
	bmcCmd.AddCommand(firmwareCmd)
	firmwareCmd.AddCommand(firmwareCheckCmd)
	firmwareCmd.AddCommand(firmwareUpgradeCmd)

	firmwareCheckCmd.Flags().BoolVar(&firmwareCheckShort, "short", false, "Only display componets with an update available")
	firmwareCheckCmd.Flags().StringVar(&firmwareCheckCatalog, "catalog", "/var/lib/grendel/repo/bmc/dell/Catalog.xml", "Path to catalog.xml from downloads.dell.com")
	firmwareCheckCmd.Flags().BoolVar(&firmwareCheckCatalogDownload, "catalog-download", false, "Auto download latest catalog from downloads.dell.com. Uses --catalog path as download location")
	firmwareCheckCmd.Flags().StringVar(&firmwareCheckFirmwareDownload, "firmware-download", "", "Path to a directory firmware will be downloaded, leaving this blank will not download the firmware. EX: /var/lib/grendel/repo/bmc")

	firmwareUpgradeCmd.Flags().StringSliceVar(&firmwareUpgradePackages, "packages", []string{}, "Path from repo endpoint to update files. (required) EX: --packages /repo/bmc/dell/idrac_fw_v7.10.0.0.EXE,/repo/bmc/dell/bios_fw_2.1.8.EXE")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradePath, "path", "", "Optional directory to packages, can be used to avoid rewriting /repo/bmc/dell for updating multiple packages EX: --path /repo/bmc/dell --packages idrac_fw_v7.10.0.0.EXE,bios_fw_2.1.8.EXE")
	firmwareUpgradeCmd.MarkFlagRequired("packages")
}

func runFirmwareCheck() error {
	pw := progress.NewWriter()
	pw.SetOutputWriter(os.Stdout)
	go pw.Render()

	pw.SetStyle(progress.StyleDefault)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Options.PercentFormat = "%4.1f%%"

	if firmwareCheckCatalogDownload {
		fmt.Printf("Downloading catalog from %s to %s\n", bmc.Dell_Catalog_Download_location, firmwareCheckCatalog)
		err := bmc.DownloadFirmware(bmc.Dell_Catalog_Download_location, firmwareCheckCatalog, "Dell LC Catalog", "", pw)
		if err != nil {
			return err
		}
	}

	job := bmc.NewJob()
	hostsFirmware, err := job.GetFirmware(hostList)
	if err != nil {
		return err
	}

	file, err := os.Open(firmwareCheckCatalog)
	if err != nil {
		return err
	}
	defer file.Close()

	nr, err := charset.NewReader(file, "utf-16")
	if err != nil {
		return err
	}

	var catalog bmc.DellCatalog

	decoder := xml.NewDecoder(nr)
	decoder.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	err = decoder.Decode(&catalog)
	if err != nil {
		return err
	}

	updateCatalog := make(map[string][]bmc.SoftwareComponents, 0)

	for _, softwareComponent := range catalog.SoftwareComponents {
		// skip OS drivers
		if softwareComponent.ComponentType == "Driver" {
			continue
		}
		for _, supportedSystem := range softwareComponent.SupportedSystems {
			for _, brand := range supportedSystem.Brand {
				for _, model := range brand.Model {
					updateCatalog[model.SystemID] = append(updateCatalog[model.SystemID], softwareComponent)
				}
			}
		}
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Host", "Component", "Current Version", "Latest Version", "Reboot Required"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:      "Host",
			AutoMerge: true,
		},
	})

	downloadFirmware := make(map[string]string, 0)
	for _, host := range hostsFirmware {
		for componentID, firmware := range host.CurrentFirmwares {
			latestFirmwares := []bmc.SoftwareComponents{}
			for _, softwareComponent := range updateCatalog[host.SystemID] {
				for _, device := range softwareComponent.SupportedDevices.Device {
					if device.ComponentID == componentID {
						latestFirmwares = append(latestFirmwares, softwareComponent)
					}
				}
			}
			latest := bmc.SoftwareComponents{}
			layout := "2006-01-02T15:04:05-07:00"
			for _, latestFirmware := range latestFirmwares {
				if latest.DateTime != "" {
					prevTime, err := time.Parse(layout, latest.DateTime)
					if err != nil {
						return err
					}
					newTime, err := time.Parse(layout, latestFirmware.DateTime)
					if err != nil {
						return err
					}
					if !prevTime.Before(newTime) {
						continue
					}
				}
				latest = latestFirmware
			}

			// --short
			if firmwareCheckShort && (firmware.Version == latest.VendorVersion || latest.VendorVersion == "") {
				continue
			}
			method := strings.ToLower(catalog.BaseLocationAccessProtocols)

			path := fmt.Sprintf("%s://%s/%s", method, catalog.BaseLocation, latest.Path)

			if firmware.Version != latest.VendorVersion {
				downloadFirmware[latest.HashMD5] = path
			}
			t.AppendRow(table.Row{
				host.Name,
				firmware.Name,
				firmware.Version,
				colorVersion(firmware.Version, latest.VendorVersion),
				latest.RebootRequired,
			}, table.RowConfig{AutoMerge: true})
		}
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()

	if firmwareCheckFirmwareDownload != "" {
		time.Sleep(time.Second)
		for i := 3; i > 0; i-- {
			fmt.Printf("Firmware auto-download will begin in: %d second(s)\n", i)
			time.Sleep(time.Second)
		}

		eg := errgroup.Group{}
		for sum, url := range downloadFirmware {
			u := url
			s := sum
			urlArr := strings.Split(url, "/")
			name := urlArr[len(urlArr)-1]
			path := fmt.Sprintf("%s/%s", firmwareCheckFirmwareDownload, name)

			eg.Go(func() error {
				return bmc.DownloadFirmware(u, path, name, s, pw)
			})
		}
		err := eg.Wait()
		if err != nil {
			return err
		}
	}

	for pw.IsRenderInProgress() {
		if pw.LengthActive() == 0 {
			pw.Stop()
		}
		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func colorVersion(v1, v2 string) string {
	if v1 != v2 {
		return text.FgHiRed.Sprint(v2)
	}
	return text.FgHiGreen.Sprint(v1)
}

func runFirmwareUpgrade() error {
	if firmwareUpgradePath != "" {
		for i, firmwareUpgradePackage := range firmwareUpgradePackages {
			firmwareUpgradePackages[i] = fmt.Sprintf("%s/%s", firmwareUpgradePath, firmwareUpgradePackage)
		}
	}

	job := bmc.NewJob()
	hosts, err := job.UpdateFirmware(hostList, firmwareUpgradePackages)
	if err != nil {
		return err
	}

	for _, host := range hosts {
		fmt.Printf("%s\t %s\t %s\n", host.Host, host.Status, host.Msg)
	}

	return nil
}
