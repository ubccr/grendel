// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"errors"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/pkg/model"
)

var log = logger.GetLogger("SWITCH")

type NetworkSwitch interface {
	GetInterfaceStatus() (model.InterfaceTable, error)
	GetMACTable() (model.MACTable, error)
	GetLLDPNeighbors() (model.LLDPNeighbors, error)
}

func NewNetworkSwitch(host *model.Host) (NetworkSwitch, error) {
	username := viper.GetString("bmc.switch_admin_username")
	password := viper.GetString("bmc.switch_admin_password")

	if username == "" || password == "" {
		log.Warn("Please set both bmc.switch_admin_username and bmc.switch_admin_password in your toml configuration file in order to query network switches")
		return nil, errors.New("failed to get switch credentials from config file")
	}

	var sw NetworkSwitch
	var err error

	bmc := host.InterfaceBMC()
	ip := ""
	if bmc != nil {
		ip = bmc.AddrString()
	}
	// TODO: automatically determine NOS
	if host.HasTags("arista") {
		sw, err = NewArista(ip, username, password)
	} else if host.HasTags("sonic") {
		sw, err = NewSonic(ip, username, password, "", true)
	} else if host.HasTags("dellos10") {
		sw, err = NewDellOS10("https://"+ip, username, password, "", true)
	} else {
		return nil, errors.New("failed to determine switch NOS")
	}

	return sw, err
}
