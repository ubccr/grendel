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
	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/ubccr/grendel/internal/bmc"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

type BmcOsPowerBody struct {
	BootOption  redfish.BootSourceOverrideTarget `json:"boot_option" description:"string of type redfish.BootSourceOverrideTarget. Common options include: None, Pxe, BiosSetup, Utilities, Diags" example:"Pxe"`
	PowerOption redfish.ResetType                `json:"power_option" description:"string of type redfish.ResetType. Common options include: On, ForceOn, ForceOff, ForceRestart, GracefulRestart, GracefulShutdown, PowerCycle" example:"PowerCycle"`
}
type BmcImportConfigurationRequest struct {
	ShutdownType string `json:"shutdown_type" description:"options include: NoReboot, Graceful, Forced" example:"Graceful"`
	File         string `json:"file" description:"template file relative to templates directory" example:"idrac-config.json.tmpl"`
}

func (h *Handler) BmcOsPower(c fuego.ContextWithBody[BmcOsPowerBody]) (model.JobMessageList, error) {
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
			Detail: "failed to parse ospower body",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.PowerControl(hostList, body.BootOption, body.PowerOption)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to submit redfish job",
		}
	}

	h.writeEvent(c.Context(), "Success", "Successfully sent OS power command to node(s)", output...)
	return output, nil
}

func (h *Handler) BmcPower(c fuego.ContextNoBody) (model.JobMessageList, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.PowerCycleBmc(hostList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to reboot bmc",
		}
	}

	h.writeEvent(c.Context(), "Success", "Successfully sent BMC power command to node(s)", output...)
	return output, nil
}

func (h *Handler) BmcSelClear(c fuego.ContextNoBody) (model.JobMessageList, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.ClearSel(hostList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to submit redfish job",
		}
	}

	h.writeEvent(c.Context(), "Success", "Successfully sent SEL Clear command to node(s)", output...)
	return output, nil
}

func (h *Handler) BmcJobList(c fuego.ContextNoBody) (model.RedfishJobList, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.GetJobs(hostList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to retrieve redfish jobs",
		}
	}

	return output, nil
}

func (h *Handler) BmcJobDelete(c fuego.ContextNoBody) (model.JobMessageList, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	jids := strings.Split(c.PathParam("jids"), ",")
	output, err := job.ClearJobs(hostList, jids)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to clear jobs",
		}
	}

	h.writeEvent(c.Context(), "Success", "Successfully sent job delete command to node(s)", output...)
	return output, nil
}

type BmcJobDeleteRequest struct {
	NodeJobList map[string][]string `json:"node_job_list" description:"Map with Node and Redfish Job IDs. Use 'JID_CLEARALL' to clear all jobs"`
}

func (h *Handler) BmcJobDeleteMany(c fuego.ContextWithBody[BmcJobDeleteRequest]) (model.JobMessageList, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse ospower body",
		}
	}

	nodeJobList := make(map[*model.Host][]string)

	for nodeName, jids := range body.NodeJobList {
		ns, err := nodeset.NewNodeSet(nodeName)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to filter nodes",
			}
		}
		hostList, err := h.DB.FindHosts(ns)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to find node",
			}
		}
		if len(hostList) != 1 {
			return nil, fuego.HTTPError{
				Err:    errors.New("nodeList length != 1"),
				Title:  "Error",
				Detail: "failed to find node",
			}
		}

		nodeJobList[hostList[0]] = jids
	}

	job := bmc.NewJob()

	output, err := job.ClearManyJobs(nodeJobList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to clear jobs",
		}
	}

	h.writeEvent(c.Context(), "Success", "Successfully sent job delete command to node(s)", output...)
	return output, nil
}

func (h *Handler) BmcQuery(c fuego.ContextNoBody) (model.RedfishSystemList, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.BmcStatus(hostList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to clear jobs",
		}
	}

	notTagged, err := nodeset.NewNodeSet("")
	if err != nil {
		return nil, err
	}

	for _, node := range hostList {
		for _, tag := range node.Tags {
			if strings.Contains(tag, "grendel:serial") {
				continue
			}
			notTagged.Add(node.Name)
		}
	}

	it := notTagged.Iterator()
	for it.Next() {
		idx := slices.IndexFunc(output, func(o model.RedfishSystem) bool { return o.Name == it.Value() })
		if idx == -1 {
			continue
		}
		job := output[idx]
		ns, err := nodeset.NewNodeSet(it.Value())
		if err != nil {
			log.Warn("failed to parse nodeset for node: ", it.Value())
			continue
		}
		if job.SerialNumber == "" {
			log.Warn("failed to get serial number for node: ", it.Value())
			continue
		}
		err = h.DB.TagHosts(ns, []string{fmt.Sprintf("grendel:serial=%s", job.SerialNumber)})
		if err != nil {
			log.Warn("failed to save updated serial number for node:", it.Value())
			continue
		}
	}

	return output, nil
}

func (h *Handler) BmcAutoConfigure(c fuego.ContextNoBody) (model.JobMessageList, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.BmcAutoConfigure(hostList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get metric reports",
		}
	}

	h.writeEvent(c.Context(), "Success", "Successfully sent BMC AutoConfigure command to node(s)", output...)
	return output, nil
}

func (h *Handler) BmcImportConfiguration(c fuego.ContextWithBody[BmcImportConfigurationRequest]) (model.JobMessageList, error) {
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
			Detail: "failed to parse ospower body",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.BmcImportConfiguration(hostList, body.ShutdownType, body.File)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get metric reports",
		}
	}

	h.writeEvent(c.Context(), "Success", "Successfully sent BMC Import Configuration to node(s)", output...)
	return output, nil
}

func (h *Handler) BmcMetricReports(c fuego.ContextNoBody) (model.RedfishMetricReportList, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.BmcGetMetricReports(hostList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get metric reports",
		}
	}

	return output, nil
}

type BmcDellInstallFromRepoRequest struct {
	ApplyUpdate       bool              `description:"ApplyUpdate false will only check for firmware upgrades. true will queue the update installs as a job."`
	IgnoreCertWarning bool              `description:"false = share needs valid HTTPS cert, true = allow invalid certs"`
	IPAddress         string            `description:"domain name or IP address of share"`
	ShareType         dell.IFRShareType `description:"type of share"`
	ShareName         string
	CatalogFile       string
	RebootNeeded      bool `description:"false = do not reboot node automatically, jobs will queue as scheduled and wait until next boot. true = reboot when needed"`
	ClearJobQueue     bool `description:"clear job queue before submitting request. ApplyUpdate must be true"`
}

func (h *Handler) BmcDellInstallFromRepo(c fuego.ContextWithBody[BmcDellInstallFromRepoRequest]) (model.JobMessageList, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse body",
		}
	}

	applyUpdate := dell.ApplyUpdateFalse
	if body.ApplyUpdate {
		applyUpdate = dell.ApplyUpdateTrue
	}
	ignoreCertWarning := dell.IgnoreCertWarningOff
	if body.IgnoreCertWarning {
		ignoreCertWarning = dell.IgnoreCertWarningOn
	}

	redfishBody := dell.InstallFromRepoBody{
		ApplyUpdate:       applyUpdate,
		IgnoreCertWarning: ignoreCertWarning,
		IPAddress:         body.IPAddress,
		ShareType:         body.ShareType,
		ShareName:         body.ShareName,
		CatalogFile:       body.CatalogFile,
		RebootNeeded:      body.RebootNeeded,
	}

	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	if body.ClearJobQueue && body.ApplyUpdate {
		jl, err := job.ClearJobs(hostList, []string{"JID_CLEARALL"})
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to clear job queue",
			}
		}
		h.writeEvent(c.Context(), "Success", "Successfully sent clear job request", jl...)
	}

	output, err := job.DellInstallFromRepo(hostList, redfishBody)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get firmware list",
		}
	}

	if body.ApplyUpdate {
		h.writeEvent(c.Context(), "Success", "Successfully sent firmware upgrade to node(s)", output...)
	}
	return output, nil
}

func (h *Handler) BmcDellGetRepoUpdateList(c fuego.ContextNoBody) (model.RedfishDellUpgradeFirmwareList, error) {
	ns, err := h.filterByNodesetAndTags(c.QueryParam("nodeset"), c.QueryParam("tags"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to filter nodes",
		}
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find nodes",
		}
	}

	job := bmc.NewJob()

	output, err := job.DellGetRepoUpdateList(hostList)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get firmware list",
		}
	}

	return output, nil
}
