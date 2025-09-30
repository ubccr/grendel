// Copyright 2021 Grendel Authors. All rights reserved.
// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	interfaceCmd = &cobra.Command{
		Use:   "interface",
		Short: "interface sub commands",
	}

	interfaceDeleteCmd = &cobra.Command{
		Use:   "delete <node name> <interface name>",
		Short: "delete an interface on a node",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			req := client.GETV1NodesFindParams{
				Nodeset: client.NewOptString(args[0]),
			}

			res, err := gc.GETV1NodesFind(context.Background(), req)
			if err != nil {
				return cmd.NewApiError(err)
			}

			if len(res) != 1 {
				return errors.New("failed to find node")
			}

			newInterfaces := make([]client.NilHostInterfacesItem, 0)
			for _, iface := range res[0].Interfaces {
				if iface.Value.Ifname.Value != args[1] {
					newInterfaces = append(newInterfaces, iface)
				}
			}

			res[0].SetInterfaces(newInterfaces)

			newJson, err := json.Marshal(res)
			if err != nil {
				return err
			}

			decodedJson := []client.NilNodeAddRequestNodeListItem{}
			err = json.Unmarshal(newJson, &decodedJson)
			if err != nil {
				return err
			}

			storeReq := client.NodeAddRequest{
				NodeList: decodedJson,
			}
			params := client.POSTV1NodesParams{}
			storeRes, err := gc.POSTV1Nodes(context.Background(), &storeReq, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(storeRes)
		},
	}

	interfaceAddBmc  bool
	interfaceAddFqdn string
	interfaceAddName string
	interfaceAddIp   string
	interfaceAddMac  string
	interfaceAddMtu  int
	interfaceAddVlan string
	interfaceAddCmd  = &cobra.Command{
		Use:   "add <node name>",
		Short: "add a new interface to a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			newIface := client.NewNilHostInterfacesItem(client.HostInterfacesItem{
				Bmc:    client.NewOptBool(interfaceAddBmc),
				Fqdn:   client.NewOptString(interfaceAddFqdn),
				Ifname: client.NewOptString(interfaceAddName),
				IP:     client.NewOptString(interfaceAddIp),
				MAC:    client.NewOptString(interfaceAddMac),
				Mtu:    client.NewOptInt(interfaceAddMtu),
				Vlan:   client.NewOptString(interfaceAddVlan),
			})

			req := client.GETV1NodesFindParams{
				Nodeset: client.NewOptString(args[0]),
			}

			res, err := gc.GETV1NodesFind(context.Background(), req)
			if err != nil {
				return cmd.NewApiError(err)
			}

			if len(res) != 1 {
				return errors.New("failed to find node")
			}

			res[0].Interfaces = append(res[0].Interfaces, newIface)

			newJson, err := json.Marshal(res)
			if err != nil {
				return err
			}

			decodedJson := []client.NilNodeAddRequestNodeListItem{}
			err = json.Unmarshal(newJson, &decodedJson)
			if err != nil {
				return err
			}

			storeReq := client.NodeAddRequest{
				NodeList: decodedJson,
			}
			params := client.POSTV1NodesParams{}
			storeRes, err := gc.POSTV1Nodes(context.Background(), &storeReq, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(storeRes)
		},
	}
)

func init() {
	interfaceAddCmd.PersistentFlags().BoolVar(&interfaceAddBmc, "bmc", false, "Bool if interface is a BMC")
	interfaceAddCmd.PersistentFlags().StringVar(&interfaceAddFqdn, "fqdn", "", "Interface Fully Qualified Domain Name")
	interfaceAddCmd.PersistentFlags().StringVar(&interfaceAddName, "name", "", "Interface name")
	interfaceAddCmd.PersistentFlags().StringVar(&interfaceAddIp, "ip", "", "Interface IP in CIDR format")
	interfaceAddCmd.PersistentFlags().StringVar(&interfaceAddMac, "mac", "", "Interface MAC address")
	interfaceAddCmd.PersistentFlags().IntVar(&interfaceAddMtu, "mtu", 1500, "Interface MTU")
	interfaceAddCmd.PersistentFlags().StringVar(&interfaceAddVlan, "vlan", "", "Interface vlan")

	nodeCmd.AddCommand(interfaceCmd)
	interfaceCmd.AddCommand(interfaceAddCmd)
	interfaceCmd.AddCommand(interfaceDeleteCmd)
}
