// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/model"
)

// BmcImportConfigBody names the config template every node's iDRAC imports
type BmcImportConfigBody struct {
	File         string `json:"file" description:"BMC config template to import, by its name in the provision templates. Each iDRAC downloads it from Grendel's provision server" example:"idrac-config.json"`
	ShutdownType string `json:"shutdown_type" description:"NoReboot queues the import until the node next reboots. Graceful and Forced reboot the node to apply it. Defaults to NoReboot" example:"NoReboot"`
}

// BmcInstallFirmwareBody points every node's iDRAC at an update catalog
type BmcInstallFirmwareBody struct {
	ApplyUpdate       bool   `json:"apply_update" description:"false only compares the installed firmware against the catalog, which firmware/updates then lists. true queues every available update as a job"`
	IPAddress         string `json:"ip_address" description:"Domain name or IP address of the share holding the catalog" example:"downloads.dell.com"`
	ShareType         string `json:"share_type" description:"HTTP, HTTPS or NFS" example:"HTTPS"`
	ShareName         string `json:"share_name" description:"Path of the catalog's directory on the share, or the NFS export" example:"/"`
	CatalogFile       string `json:"catalog_file" description:"Catalog file name on the share. The iDRAC defaults to Catalog.xml" example:"Catalog.xml"`
	IgnoreCertWarning bool   `json:"ignore_cert_warning" description:"Accept an HTTPS share whose certificate does not validate"`
	Reboot            bool   `json:"reboot" description:"Reboot the node when an update needs it. Otherwise the updates wait as scheduled jobs until the node next reboots"`
	ClearJobs         bool   `json:"clear_jobs" description:"Clear every job in the queue first. Requires apply_update"`
}

// BmcAutoConfigure sets iDRAC(s) to auto configure once on their next boot
func (h *Handler) BmcAutoConfigure(c fuego.ContextNoBody) (model.BmcResponse[model.BmcAck], error) {
	return runAck(h, c, "bmc configure auto", func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.AutoConfigure(ctx, t)
	})
}

// BmcImportConfig has iDRAC(s) download and import a config template from the provision server
func (h *Handler) BmcImportConfig(c fuego.ContextWithBody[BmcImportConfigBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, badRequest(err, "failed to parse the import request")
	}

	shutdown, err := shutdownType(body.ShutdownType)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}
	// the provision server serves bmc templates at /boot/<token>/bmc/<name>, one path segment
	if body.File == "" || strings.Contains(body.File, "/") {
		return model.BmcResponse[model.BmcAck]{}, badRequest(fmt.Errorf("invalid file %q", body.File), "file must be a template name with no path")
	}

	share, err := importShare(viper.GetViper())
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to work out the address bmcs fetch the config from"}
	}

	hosts, err := h.bmcSelect(c)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	req := redfish.ImportConfigRequest{
		Body: dell.ImportSystemConfigurationBody{
			HostPowerState: dell.OnISCHostPowerState,
			ShutdownType:   shutdown,
			ShareParameters: dell.ShareParameters{
				Target:                   "ALL",
				ShareType:                share.shareType,
				IPAddress:                share.ip,
				FileName:                 body.File,
				IgnoreCertificateWarning: share.ignoreCert,
			},
		},
		PortNumber: share.port,
		ShareNames: shareNames(hosts),
	}

	return runAck(h, c, fmt.Sprintf("bmc configure import %s, shutdown %s", body.File, shutdown), func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.ImportConfig(ctx, t, req)
	})
}

// BmcFirmwareUpdates lists the firmware updates iDRAC(s) found in their last catalog comparison
func (h *Handler) BmcFirmwareUpdates(c fuego.ContextNoBody) (model.BmcResponse[model.BmcFirmwareUpdates], error) {
	return runBmc(h, c, "bmc firmware updates", false,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.FirmwareUpdates] {
			return h.Redfish.FirmwareUpdates(ctx, t)
		},
		firmwareUpdates,
	)
}

// BmcInstallFirmware has iDRAC(s) compare their firmware against a catalog, and queue the updates when asked to
func (h *Handler) BmcInstallFirmware(c fuego.ContextWithBody[BmcInstallFirmwareBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, badRequest(err, "failed to parse the firmware request")
	}

	req, err := installFirmwareRequest(body)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	operation := "bmc firmware check against " + body.IPAddress
	if body.ApplyUpdate {
		operation = "bmc firmware install from " + body.IPAddress
	}

	// a comparison changes nothing on the node, so only an install is recorded in the event log
	return runBmc(h, c, operation, body.ApplyUpdate,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
			return h.Redfish.InstallFirmware(ctx, t, req)
		},
		func(*redfish.Ack) *model.BmcAck { return nil },
	)
}

func shutdownType(raw string) (dell.ISCShutdownType, error) {
	switch raw {
	case "", string(dell.NoRebootISCShutdownType):
		return dell.NoRebootISCShutdownType, nil
	case string(dell.GracefulISCShutdownType):
		return dell.GracefulISCShutdownType, nil
	case string(dell.ForcedISCShutdownType):
		return dell.ForcedISCShutdownType, nil
	}
	return "", badRequest(fmt.Errorf("unknown shutdown type %q", raw), "shutdown_type must be NoReboot, Graceful or Forced")
}

type configShare struct {
	ip         string
	port       string
	shareType  dell.ISCShareType
	ignoreCert dell.ISCIgnoreCertificateWarning
}

// importShare is where iDRACs fetch an imported config from, the provision server. bmc.config_share_ip wins, then the address provision listens on, then this machine's first external address when provision listens on every interface. The port is always provision's, and is only set when it is not the share type's default
func importShare(v *viper.Viper) (configShare, error) {
	share := configShare{
		shareType:  dell.HTTPISCShareType,
		ignoreCert: dell.DisabledISCIgnoreCertificateWarning,
	}
	defaultPort := "80"
	if v.IsSet("provision.cert") {
		share.shareType = dell.HTTPSISCShareType
		defaultPort = "443"
	}
	if v.GetString("bmc.config_ignore_certificate_warning") == "Enabled" {
		share.ignoreCert = dell.EnabledISCIgnoreCertificateWarning
	}

	listen, port, err := net.SplitHostPort(v.GetString("provision.listen"))
	if err != nil {
		return share, fmt.Errorf("parsing provision.listen: %w", err)
	}
	if port != defaultPort {
		share.port = port
	}

	switch ip := v.GetString("bmc.config_share_ip"); {
	case ip != "":
		share.ip = ip
	case listen != "" && listen != "0.0.0.0":
		share.ip = listen
	default:
		external, err := util.GetFirstExternalIPFromInterfaces()
		if err != nil {
			return share, err
		}
		share.ip = external.String()
	}

	return share, nil
}

// shareNames gives every node the provision path its iDRAC downloads the config from, which carries a boot token for that node. A node whose token cannot be made is left out, and fails on its own
func shareNames(hosts model.HostList) map[string]string {
	names := make(map[string]string, len(hosts))
	for _, host := range hosts {
		if host == nil {
			continue
		}
		nic := host.InterfaceBMC()
		if nic == nil {
			continue
		}

		token, err := model.NewBootToken(host.UID.String(), nic.MAC.String())
		if err != nil {
			log.Warnf("failed to create a boot token for %s: %s", host.Name, err)
			continue
		}
		names[host.Name] = "/boot/" + token + "/bmc"
	}
	return names
}

var firmwareShareTypes = []dell.IFRShareType{dell.HTTPIFRShareType, dell.HTTPSIFRShareType, dell.NFSIFRShareType}

// installFirmwareRequest validates the body before any node is contacted
func installFirmwareRequest(body BmcInstallFirmwareBody) (redfish.InstallFirmwareRequest, error) {
	shareType := dell.IFRShareType(strings.ToUpper(body.ShareType))
	if !slices.Contains(firmwareShareTypes, shareType) {
		return redfish.InstallFirmwareRequest{}, badRequest(fmt.Errorf("unknown share type %q", body.ShareType), "share_type must be HTTP, HTTPS or NFS")
	}
	if body.IPAddress == "" {
		return redfish.InstallFirmwareRequest{}, badRequest(errors.New("no ip_address given"), "ip_address is required")
	}
	if body.ShareName == "" {
		return redfish.InstallFirmwareRequest{}, badRequest(errors.New("no share_name given"), "share_name is required")
	}
	if body.ClearJobs && !body.ApplyUpdate {
		return redfish.InstallFirmwareRequest{}, badRequest(errors.New("clear_jobs without apply_update"), "clear_jobs requires apply_update")
	}

	apply := dell.ApplyUpdateFalse
	if body.ApplyUpdate {
		apply = dell.ApplyUpdateTrue
	}
	ignoreCert := dell.IgnoreCertWarningOff
	if body.IgnoreCertWarning {
		ignoreCert = dell.IgnoreCertWarningOn
	}

	return redfish.InstallFirmwareRequest{
		Body: dell.InstallFromRepoBody{
			ApplyUpdate:       apply,
			IPAddress:         body.IPAddress,
			ShareType:         shareType,
			ShareName:         body.ShareName,
			CatalogFile:       body.CatalogFile,
			IgnoreCertWarning: ignoreCert,
			RebootNeeded:      body.Reboot,
		},
		ClearJobs: body.ClearJobs,
	}, nil
}

// rebootRank orders reboot types by how disruptive they are
var rebootRank = map[string]int{"NONE": 0, "IDRAC": 1, "HOST": 2}

func firmwareUpdates(data *redfish.FirmwareUpdates) *model.BmcFirmwareUpdates {
	if data == nil {
		return nil
	}

	out := &model.BmcFirmwareUpdates{Updates: make([]model.BmcFirmwareUpdate, 0, len(data.Updates)), RebootType: "NONE"}
	for _, u := range data.Updates {
		out.Updates = append(out.Updates, model.BmcFirmwareUpdate{
			Name:             u.Name,
			DisplayName:      u.DisplayName,
			ComponentType:    u.ComponentType,
			InstalledVersion: u.InstalledVersion,
			PackageVersion:   u.PackageVersion,
			Criticality:      u.Criticality,
			RebootType:       u.RebootType,
			PackagePath:      u.PackagePath,
		})
		if rebootRank[u.RebootType] > rebootRank[out.RebootType] {
			out.RebootType = u.RebootType
		}
	}

	return out
}
