// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"fmt"

	"github.com/ubccr/grendel/internal/config"
)

const (
	endpointPrefix             string = "boot"
	endpointRepo                      = "repo"
	endpointComplete                  = "complete"
	endpointIPXE                      = "ipxe"
	endpointKickstart                 = "kickstart"
	endpointKernel                    = "file/kernel"
	endpointLiveImage                 = "file/liveimg"
	endpointRootFS                    = "file/rootfs"
	endpointInitrd                    = "file/initrd"
	endpointCloudInit                 = "cloud-init/"
	endpointUserData                  = "cloud-init/user-data"
	endpointMetaData                  = "cloud-init/meta-data"
	endpointVendorData                = "cloud-init/vendor-data"
	endpointIgnition                  = "pxe-config.ign"
	endpointProvision                 = "provision/"
	endpointProxmox                   = "proxmox"
	endpointNetBoxRenderConfig        = "netbox/render-config"
)

type Endpoints struct {
	host  string
	token string
}

func NewEndpoints(host, token string) *Endpoints {
	return &Endpoints{host: host, token: token}
}

func (e *Endpoints) BootFileURL() string {
	return fmt.Sprintf("tftp://%s/%s", e.host, e.token)
}

func (e *Endpoints) RepoURL() string {
	return fmt.Sprintf("%s/%s", e.BaseURL(), endpointRepo)
}

func (e *Endpoints) BaseURL() string {
	host := e.host
	if config.ProvisionHostname != "" {
		host = config.ProvisionHostname
	}

	baseURL := fmt.Sprintf("%s://%s", config.ProvisionScheme, host)
	if config.ProvisionAddr.Port() != 80 && config.ProvisionAddr.Port() != 443 {
		baseURL += fmt.Sprintf(":%d", config.ProvisionAddr.Port())
	}

	return baseURL
}

func (e *Endpoints) provisionURL(endpoint string) string {
	return fmt.Sprintf("%s/%s/%s/%s", e.BaseURL(), endpointPrefix, e.token, endpoint)
}

func (e *Endpoints) CompleteURL() string {
	return e.provisionURL(endpointComplete)
}

func (e *Endpoints) IpxeURL() string {
	return e.provisionURL(endpointIPXE)
}

func (e *Endpoints) KickstartURL() string {
	return e.provisionURL(endpointKickstart)
}

func (e *Endpoints) KickstartURLParts() (string, string) {
	return e.provisionURL(""), endpointKickstart
}

func (e *Endpoints) KernelURL() string {
	return e.provisionURL(endpointKernel)
}

func (e *Endpoints) LiveImageURL() string {
	return e.provisionURL(endpointLiveImage)
}

func (e *Endpoints) RootFSURL() string {
	return e.provisionURL(endpointRootFS)
}

func (e *Endpoints) InitrdURL(index int) string {
	return e.provisionURL(fmt.Sprintf("%s-%d", endpointInitrd, index))
}

func (e *Endpoints) CloudInitURL() string {
	return e.provisionURL(endpointCloudInit)
}

func (e *Endpoints) UserDataURL() string {
	return e.provisionURL(endpointUserData)
}

func (e *Endpoints) MetaDataURL() string {
	return e.provisionURL(endpointMetaData)
}

func (e *Endpoints) VendorDataURL() string {
	return e.provisionURL(endpointVendorData)
}

func (e *Endpoints) IgnitionURL() string {
	return e.provisionURL(endpointIgnition)
}

func (e *Endpoints) ProvisionURL(name string) string {
	return e.provisionURL(endpointProvision + name)
}

func (e *Endpoints) ProxmoxURL() string {
	return e.provisionURL(endpointProxmox)
}

func (e *Endpoints) NetBoxRenderConfigURL() string {
	return e.provisionURL(endpointNetBoxRenderConfig)
}
