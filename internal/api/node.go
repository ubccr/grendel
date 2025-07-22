// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

type NodeProvisionRequest struct {
	Provision bool `json:"provision"`
}

type NodeTagsRequest struct {
	Tags string `json:"tags" description:"comma separated list of tags" example:"a01,test"`
}

type NodeAddRequest struct {
	NodeList model.HostList `json:"node_list"`
}

type NodeBootTokenResponse struct {
	Nodes []struct {
		Name  string `json:"name"`
		Token string `json:"token"`
	} `json:"nodes"`
}
type NodeBootImageRequest struct {
	Image string `json:"image"`
}

func (h *Handler) NodeAdd(c fuego.ContextWithBody[NodeAddRequest]) (*GenericResponse, error) {
	body, err := c.Body()
	if err != nil {
		// TODO: add error passthrough on non sensitive endpoints
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: fmt.Sprintf("failed to parse body: %s", err.Error()),
		}
	}

	err = h.DB.StoreHosts(body.NodeList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: fmt.Sprintf("failed to store node(s): %s", err.Error()),
		}
	}

	ns, err := body.NodeList.ToNodeSet()
	if err == nil {
		h.writeEvent(c.Context(), "Success", fmt.Sprintf("Successfully saved node(s): %s", ns.String()))
	}

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully added node(s)",
		Changed: len(body.NodeList),
	}, nil
}

func (h *Handler) NodeList(c fuego.ContextNoBody) (model.HostList, error) {
	NodeList, err := h.DB.Hosts()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get nodes",
		}
	}

	return NodeList, nil
}

func (h *Handler) NodeFind(c fuego.ContextNoBody) (model.HostList, error) {
	// TODO: implement a native DB func to handle this?
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	var NodeList model.HostList
	if ns.Len() == 0 {
		NodeList, err = h.DB.Hosts()
	} else {
		NodeList, err = h.DB.FindHosts(ns)
	}
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	return NodeList, nil
}

func (h *Handler) NodeDelete(c fuego.ContextNoBody) (*GenericResponse, error) {
	// TODO: implement a native DB func to handle this?
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	err = h.DB.DeleteHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to delete nodes",
		}
	}

	h.writeEvent(c.Context(), "Success", fmt.Sprintf("Successfully deleted node(s): %s", ns.String()))

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully deleted node(s)",
		Changed: ns.Len(),
	}, nil
}

func (h *Handler) NodeProvision(c fuego.ContextWithBody[NodeProvisionRequest]) (*GenericResponse, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse body",
		}
	}
	err = h.DB.ProvisionHosts(ns, body.Provision)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to change provision on node(s)",
		}
	}
	return &GenericResponse{
		Title:   "Success",
		Detail:  fmt.Sprintf("successfully changed node(s) provision to %t", body.Provision),
		Changed: ns.Len(),
	}, nil
}

func (h *Handler) NodeTags(c fuego.ContextWithBody[NodeTagsRequest]) (*GenericResponse, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: "failed to filter nodes",
		}
	}
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: "failed to parse body",
		}
	}

	if body.Tags == "" {
		return nil, fuego.HTTPError{
			Title:  "Error",
			Detail: "invalid list of update tags",
		}
	}

	tags := strings.Split(body.Tags, ",")
	msg := ""

	if c.PathParam("action") == "add" {
		msg = "successfully added tag(s) to node(s)"
		err = h.DB.TagHosts(ns, tags)
	} else if c.PathParam("action") == "remove" {
		msg = "successfully removed tag(s) from node(s)"
		err = h.DB.UntagHosts(ns, tags)
	}

	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to update node(s) tags",
		}
	}

	return &GenericResponse{
		Title:   "Success",
		Detail:  msg,
		Changed: ns.Len(),
	}, nil
}

func (h *Handler) NodeBootToken(c fuego.ContextNoBody) (*NodeBootTokenResponse, error) {
	iface := c.PathParam("interface")

	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	nodeList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}
	if len(nodeList) == 0 {
		return nil, fuego.HTTPError{
			Err:    errors.New("no hosts found with nodeset"),
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	var output NodeBootTokenResponse

	for _, node := range nodeList {
		if len(node.Interfaces) == 0 {
			continue
		}

		var nic *model.NetInterface
		switch iface {
		case "boot":
			nic = node.BootInterface()
		case "bmc":
			nic = node.InterfaceBMC()
		default:
			nic = node.BootInterface()
		}

		token, err := model.NewBootToken(node.UID.String(), nic.MAC.String())
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to generate boot token",
			}
		}

		output.Nodes = append(output.Nodes, struct {
			Name  string `json:"name"`
			Token string `json:"token"`
		}{Name: node.Name, Token: token})
	}

	return &output, nil
}

func (h *Handler) NodeBootImage(c fuego.ContextWithBody[NodeBootImageRequest]) (*GenericResponse, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse body",
		}
	}

	err = h.DB.SetBootImage(ns, body.Image)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to update boot images",
		}
	}

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully updated node(s) boot image",
		Changed: ns.Len(),
	}, nil
}

func (h *Handler) filterByNodesetAndTags(f1, f2 string) (*nodeset.NodeSet, error) {
	queryNs, err := nodeset.NewNodeSet(f1)
	if err != nil {
		return nil, err
	}
	tagsNs, _ := nodeset.NewNodeSet("")
	if f2 != "" {
		tags := strings.Split(f2, ",")
		tagsNs, err = h.DB.FindTags(tags)
		if err != nil {
			return nil, err
		}
	}

	compare1 := queryNs.Iterator().StringSlice()
	compare2 := tagsNs.Iterator().StringSlice()

	combined := []string{}
	for _, node := range compare1 {
		if slices.Contains(compare2, node) {
			combined = append(combined, node)
		}
	}

	if len(compare1) == 0 {
		combined = compare2
	}
	if len(compare2) == 0 {
		combined = compare1
	}

	return nodeset.NewNodeSet(strings.Join(combined, ","))
}
