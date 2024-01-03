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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kgrvamsi/redfishapi"
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type Redfish struct {
	config gofish.ClientConfig
	client *gofish.APIClient
}

var (
	powerCycleTypeOrder = []string{
		"PowerCycle",
		"GracefulRestart",
		"ForceRestart",
	}
	powerOnTypeOrder = []string{
		"On",
		"ForceOn",
	}
	powerOffTypeOrder = []string{
		"ForceOff",
		"GracefulShutdown",
	}
)

func NewRedfish(endpoint, user, pass string, insecure bool) (*Redfish, error) {
	config := gofish.ClientConfig{
		Endpoint: endpoint,
		Username: user,
		Password: pass,
		Insecure: insecure,
	}

	fish, err := gofish.Connect(config)
	if err != nil {
		return nil, err
	}

	return &Redfish{config: config, client: fish}, nil
}

func (r *Redfish) Logout() {
	r.client.Logout()
}
func IdracAutoConfigure(ip string) error {

	user := viper.GetString("bmc.user")
	pass := viper.GetString("bmc.password")

	c := redfishapi.NewIloClient(fmt.Sprintf("https://%s", ip), user, pass)

	// TODO: replace redfish library or submit PR to fix "k.MessageExtendedInfo[0].Message"
	// License error response has MessageExtendedInfo nested in a error struct
	attr, err := c.GetIDRACAttrDell()
	if err != nil {
		return err
	}
	if attr.NIC_1_AutoConfig == "" {
		return errors.New("NIC_1_AutoConfig attribute not found. iDRAC is likely missing the required license")
	}

	type attributes struct {
		NIC1AutoConfig string `json:"NIC.1.AutoConfig"`
	}
	type body struct {
		Attributes attributes `json:"Attributes"`
	}
	b, _ := json.Marshal(body{
		Attributes: attributes{
			NIC1AutoConfig: "Enable Once",
		},
	})

	_, err = c.SetAttributesDell("idrac", b)
	if err != nil {
		// Try default creds
		c.Username = "root"
		c.Password = "calvin"
		_, err = c.SetAttributesDell("idrac", b)
	}
	return err

}
func RebootHost(ip string, bootOverride redfish.Boot) error {
	config := gofish.ClientConfig{
		Endpoint: fmt.Sprintf("https://%s", ip),
		Username: viper.GetString("bmc.user"),
		Password: viper.GetString("bmc.password"),
		Insecure: true,
	}

	c, err := gofish.Connect(config)
	if err != nil {
		return err
	}
	defer c.Logout()

	// Attached the client to service root
	service := c.Service
	ss, err := service.Systems()
	if err != nil {
		return err
	}

	for _, system := range ss {
		err := system.SetBoot(bootOverride)
		if err != nil {
			return err
		}
		err = system.Reset(redfish.ForceRestartResetType)
		if err != nil {
			return err
		}
	}
	return nil
}
func IdracImportSytemConfig(ip string, path string, file string) error {

	user := viper.GetString("bmc.user")
	pass := viper.GetString("bmc.password")

	c := redfishapi.NewIloClient(fmt.Sprintf("https://%s", ip), user, pass)

	type shareParameters struct {
		Target     []string `json:"Target"`
		ShareType  string   `json:"ShareType"`
		IPAddress  string   `json:"IPAddress"`
		FileName   string   `json:"FileName"`
		ShareName  string   `json:"ShareName"`
		PortNumber string   `json:"PortNumber"`
	}
	type body struct {
		HostPowerState  string `json:"HostPowerState"`
		ShutdownType    string `json:"ShutdownType"`
		ImportBuffer    string `json:"ImportBuffer"`
		ShareParameters shareParameters
	}

	b, _ := json.Marshal(body{
		HostPowerState: "On",
		ShutdownType:   "Graceful",
		ShareParameters: shareParameters{
			Target:     []string{"ALL"},
			ShareType:  viper.GetString("bmc.config_share_type"),
			IPAddress:  viper.GetString("bmc.config_share_ip"),
			PortNumber: viper.GetString("bmc.config_share_port"),
			FileName:   file,
			ShareName:  path,
		},
	})
	_, err := c.ImportConfigDell(b)

	if err != nil && err.Error() == "Unauthorized" {
		// Try default creds
		c.Username = "root"
		c.Password = "calvin"
		_, err = c.ImportConfigDell(b)
	}

	return err

}

func (r *Redfish) powerReset(resetTypeOrder []string) error {
	service := r.client.Service
	ss, err := service.Systems()
	if err != nil {
		return err
	}

	// XXX Only reset the first supported system?
	for _, system := range ss {
		for _, resetType := range resetTypeOrder {
			for _, rt := range system.SupportedResetTypes {
				if resetType == string(rt) {
					err = system.Reset(rt)
					if err != nil {
						return err
					}
					return nil
				}
			}
		}
	}

	return errors.New("Failed to find a supported reset type")
}

func (r *Redfish) PowerCycle() error {
	return r.powerReset(powerCycleTypeOrder)
}

func (r *Redfish) PowerOn() error {
	return r.powerReset(powerOnTypeOrder)
}

func (r *Redfish) PowerOff() error {
	return r.powerReset(powerOffTypeOrder)
}

func (r *Redfish) EnablePXE() error {
	service := r.client.Service
	ss, err := service.Systems()
	if err != nil {
		return err
	}

	bootOverride := redfish.Boot{
		BootSourceOverrideTarget:  redfish.PxeBootSourceOverrideTarget,
		BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
	}

	for _, system := range ss {
		err := system.SetBoot(bootOverride)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Redfish) GetSystem() (*System, error) {
	service := r.client.Service
	ss, err := service.Systems()
	if err != nil {
		return nil, err
	}

	if len(ss) == 0 {
		return nil, errors.New("Failed to find system")
	}

	sys := ss[0]

	system := &System{
		Name:           sys.HostName,
		BIOSVersion:    sys.BIOSVersion,
		SerialNumber:   sys.SKU,
		Manufacturer:   sys.Manufacturer,
		PowerStatus:    string(sys.PowerState),
		Health:         string(sys.Status.Health),
		TotalMemory:    sys.MemorySummary.TotalSystemMemoryGiB,
		ProcessorCount: sys.ProcessorSummary.LogicalProcessorCount,
		BootNext:       sys.Boot.BootNext,
		BootOrder:      sys.Boot.BootOrder,
	}

	return system, nil
}
