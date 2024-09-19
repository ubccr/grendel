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

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/bmc"
	"golang.org/x/net/html/charset"
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
	firmwareCheckCatalog string
	firmwareCheckShort   bool

	firmwareUpgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade firmware packages",
		Long:  `Upgrade firmware packages via Redfish Simple Update. This may take a while!`,
		RunE: func(command *cobra.Command, args []string) error {
			return runFirmwareUpgrade()
		},
	}
	firmwareUpgradePaths []string
	// firmwareUpgradeReboot bool
)

func init() {
	bmcCmd.AddCommand(firmwareCmd)
	firmwareCmd.AddCommand(firmwareCheckCmd)
	firmwareCmd.AddCommand(firmwareUpgradeCmd)

	firmwareCheckCmd.Flags().StringVar(&firmwareCheckCatalog, "catalog", "", "Path to catalog.xml from downloads.dell.com")
	firmwareCheckCmd.MarkFlagRequired("catalog")
	firmwareCheckCmd.Flags().BoolVar(&firmwareCheckShort, "short", false, "Only display componets with an update available")

	firmwareUpgradeCmd.Flags().StringSliceVar(&firmwareUpgradePaths, "paths", []string{}, "Path from /repo directory to update files. (required) EX: --paths /repo/bmc/dell/idrac_fw_v7.10.0.0.EXE,/repo/bmc/dell/bios_fw_2.1.8.EXE")
	firmwareUpgradeCmd.MarkFlagRequired("paths")
	// firmwareUpgradeCmd.Flags().BoolVar(&firmwareUpgradeReboot, "reboot", false, "Reboot host automatically if required by Firmware upgrade")
}

type Catalog struct {
	BaseLocation                string               `xml:"baseLocation,attr"`
	BaseLocationAccessProtocols string               `xml:"baseLocationAccessProtocols,attr"`
	DateTime                    string               `xml:"dateTime,attr"`
	Version                     string               `xml:"version,attr"`
	SoftwareComponents          []SoftwareComponents `xml:"SoftwareComponent"`
}
type SoftwareComponents struct {
	Name            string `xml:"Name>Display"`
	DateTime        string `xml:"dateTime,attr"`
	DellVersion     string `xml:"dellVersion,attr"`
	Path            string `xml:"path,attr"`
	RebootRequired  string `xml:"rebootRequired,attr"`
	ReleaseDate     string `xml:"releaseDate,attr"`
	Size            string `xml:"size,attr"`
	VendorVersion   string `xml:"vendorVersion,attr"`
	ComponentType   string `xml:"ComponentType>Display"`
	Description     string `xml:"Description>Display"`
	LUCategory      string `xml:"LUCategory>Display"`
	Category        string `xml:"Category>Display"`
	RevisionHistory string `xml:"RevisionHistory>Display"`
	Criticality     string `xml:"Criticality>Display"`
	ImportantInfo   struct {
		Info string `xml:"Display"`
		URL  string `xml:"URL,attr"`
	}
	SupportedDevices struct {
		Device []struct {
			Name        string `xml:"Display"`
			ComponentID string `xml:"componentID,attr"`
			Embedded    string `xml:"embedded,attr"`
			PCIInfo     struct {
				DeviceID    string `xml:"deviceID,attr"`
				SubDeviceID string `xml:"subDeviceID,attr"`
				VendorID    string `xml:"vendorID,attr"`
				SubVendorID string `xml:"subVendorID,attr"`
			}
		}
	}
	SupportedSystems []struct {
		Brand []struct {
			Key    string `xml:"key,attr"`
			Prefix string `xml:"prefix,attr"`
			Name   string `xml:"Display"`
			Model  []struct {
				SystemID     string `xml:"systemID,attr"`
				SystemIDType string `xml:"systemIDType,attr"`
				Name         string `xml:"Display"`
			}
		}
	}
}

func runFirmwareCheck() error {
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

	var catalog Catalog

	decoder := xml.NewDecoder(nr)
	decoder.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	err = decoder.Decode(&catalog)
	if err != nil {
		return err
	}

	updateCatalog := make(map[string][]SoftwareComponents, 0)

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
	t2 := table.NewWriter()
	t2.SetOutputMirror(os.Stdout)
	t2.AppendHeader(table.Row{"Component", "Path"})
	t2.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:      "Component",
			AutoMerge: true,
		},
		{
			Name:      "Path",
			AutoMerge: true,
		},
	})

	for _, host := range hostsFirmware {
		for componentID, firmware := range host.CurrentFirmwares {
			latestFirmwares := []SoftwareComponents{}
			for _, softwareComponent := range updateCatalog[host.SystemID] {
				for _, device := range softwareComponent.SupportedDevices.Device {
					if device.ComponentID == componentID {
						latestFirmwares = append(latestFirmwares, softwareComponent)
					}
				}
			}
			latest := SoftwareComponents{}
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

			t.AppendRow(table.Row{
				host.Name,
				firmware.Name,
				firmware.Version,
				colorVersion(firmware.Version, latest.VendorVersion),
				latest.RebootRequired,
			}, table.RowConfig{AutoMerge: true})
			t2.AppendRow(table.Row{
				firmware.Name,
				path,
			}, table.RowConfig{AutoMerge: true})
			t2.AppendSeparator()
		}
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
	t2.SetStyle(table.StyleLight)
	t2.SortBy([]table.SortBy{
		{Name: "Component", Mode: table.Asc},
	})
	t2.Render()

	return nil
}

func colorVersion(v1, v2 string) string {
	if v1 != v2 {
		return text.FgHiRed.Sprint(v2)
	}
	return text.FgHiGreen.Sprint(v1)
}

func runFirmwareUpgrade() error {
	job := bmc.NewJob()
	hosts, err := job.UpdateFirmware(hostList, firmwareUpgradePaths)
	if err != nil {
		return err
	}

	for _, host := range hosts {
		fmt.Printf("%s\t %s\t %s\n", host.Host, host.Status, host.Msg)
	}

	return nil
}
