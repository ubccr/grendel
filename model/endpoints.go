package model

import (
	"fmt"
)

const (
	endpointPrefix     string = "boot"
	endpointRepo              = "repo"
	endpointComplete          = "complete"
	endpointIPXE              = "ipxe"
	endpointKickstart         = "kickstart"
	endpointKernel            = "file/kernel"
	endpointLiveImage         = "file/liveimg"
	endpointRootFS            = "file/rootfs"
	endpointInitrd            = "file/initrd"
	endpointCloudInit         = "cloud-init/"
	endpointUserData          = "cloud-init/user-data"
	endpointMetaData          = "cloud-init/meta-data"
	endpointVendorData        = "cloud-init/vendor-data"
	endpointIgnition          = "pxe-config.ign"
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
	if ProvisionHostname != "" {
		host = ProvisionHostname
	}

	baseURL := fmt.Sprintf("%s://%s", ProvisionScheme, host)
	if ProvisionAddr.Port() != 80 && ProvisionAddr.Port() != 443 {
		baseURL += fmt.Sprintf(":%d", ProvisionAddr.Port())
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

func (e *Endpoints) CloudInit() string {
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
