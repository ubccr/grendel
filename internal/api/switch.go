package api

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/facette/natsort"
	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/tors"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

func (h *Handler) SwitchGetLLDP(c fuego.ContextNoBody) (*model.SwitchLLDPList, error) {
	filter := c.QueryParam("ports")
	csf := []string{}
	if filter != "" {
		csf = strings.Split(filter, ",")

	}

	// ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	// if err != nil {
	// 	return nil, fuego.HTTPError{
	// 		Err:    err,
	// 		Title:  "Error",
	// 		Detail: "failed to filter nodes",
	// 	}
	// }
	ns, err := nodeset.NewNodeSet(c.PathParam("nodeset"))
	if err != nil || ns.Len() > 1 {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to create nodeset. Only one switch can be queried at a time",
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
	if len(nodeList) != 1 {
		return nil, fuego.HTTPError{
			Err:    errors.New("query requires filtering to one node"),
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	netSwitch, err := tors.NewNetworkSwitch(nodeList[0])
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: err.Error(),
		}
	}

	res, err := netSwitch.GetLLDPNeighbors()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to retrieve lldp neighbors",
		}
	}

	filteredList := make(model.LLDPNeighbors)
	fmt.Println(csf, len(csf))
	if len(csf) > 0 {
		for _, port := range csf {
			filteredList[port] = res[port]
		}
	} else {
		filteredList = res
	}

	resSlice := make(model.SwitchLLDPList, 0)

	for _, v := range filteredList {
		resSlice = append(resSlice, *v)
	}

	slices.SortFunc(resSlice, func(a, b model.LLDP) int {
		if natsort.Compare(a.PortName, b.PortName) {
			return -1
		}

		return 1
	})

	return &resSlice, nil
}
