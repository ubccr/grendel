// Code generated by ogen, DO NOT EDIT.

package client

import (
	"fmt"
	"time"

	"github.com/go-faster/jx"
)

// SetFake set fake values.
func (s *AuthRequest) SetFake() {
	{
		{
			s.Password.SetFake()
		}
	}
	{
		{
			s.Username.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *AuthResponse) SetFake() {
	{
		{
			s.Expire.SetFake()
		}
	}
	{
		{
			s.Role.SetFake()
		}
	}
	{
		{
			s.Token.SetFake()
		}
	}
	{
		{
			s.Username.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *AuthSignupRequest) SetFake() {
	{
		{
			s.Password = "string"
		}
	}
	{
		{
			s.Username = "string"
		}
	}
}

// SetFake set fake values.
func (s *AuthTokenReponse) SetFake() {
	{
		{
			s.Token.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *AuthTokenRequest) SetFake() {
	{
		{
			s.Expire.SetFake()
		}
	}
	{
		{
			s.Role.SetFake()
		}
	}
	{
		{
			s.Username.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *BmcImportConfigurationRequest) SetFake() {
	{
		{
			s.File.SetFake()
		}
	}
	{
		{
			s.ShutdownType.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *BmcOsPowerBody) SetFake() {
	{
		{
			s.BootOption.SetFake()
		}
	}
	{
		{
			s.PowerOption.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *BootImage) SetFake() {
	{
		{
			s.Butane.SetFake()
		}
	}
	{
		{
			s.Cmdline.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Initrd = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Initrd = append(s.Initrd, elem)
			}
		}
	}
	{
		{
			s.Kernel = "string"
		}
	}
	{
		{
			s.Liveimg.SetFake()
		}
	}
	{
		{
			s.Name = "string"
		}
	}
	{
		{
			s.ProvisionTemplate.SetFake()
		}
	}
	{
		{
			s.ProvisionTemplates.SetFake()
		}
	}
	{
		{
			s.UID.SetFake()
		}
	}
	{
		{
			s.UserData.SetFake()
		}
	}
	{
		{
			s.Verify.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *BootImageAddRequest) SetFake() {
	{
		{
			s.BootImages = nil
			for i := 0; i < 0; i++ {
				var elem NilBootImageAddRequestBootImagesItem
				{
					elem.SetFake()
				}
				s.BootImages = append(s.BootImages, elem)
			}
		}
	}
}

// SetFake set fake values.
func (s *BootImageAddRequestBootImagesItem) SetFake() {
	{
		{
			s.Butane.SetFake()
		}
	}
	{
		{
			s.Cmdline.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Initrd = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Initrd = append(s.Initrd, elem)
			}
		}
	}
	{
		{
			s.Kernel.SetFake()
		}
	}
	{
		{
			s.Liveimg.SetFake()
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.ProvisionTemplate.SetFake()
		}
	}
	{
		{
			s.ProvisionTemplates.SetFake()
		}
	}
	{
		{
			s.UID.SetFake()
		}
	}
	{
		{
			s.UserData.SetFake()
		}
	}
	{
		{
			s.Verify.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *BootImageAddRequestBootImagesItemProvisionTemplates) SetFake() {
	var (
		elem NilString
		m    map[string]NilString = s.init()
	)
	for i := 0; i < 0; i++ {
		m[fmt.Sprintf("fake%d", i)] = elem
	}
}

// SetFake set fake values.
func (s *BootImageProvisionTemplates) SetFake() {
	var (
		elem NilString
		m    map[string]NilString = s.init()
	)
	for i := 0; i < 0; i++ {
		m[fmt.Sprintf("fake%d", i)] = elem
	}
}

// SetFake set fake values.
func (s *DataDump) SetFake() {
	{
		{
			s.Hosts = nil
			for i := 0; i < 0; i++ {
				var elem NilDataDumpHostsItem
				{
					elem.SetFake()
				}
				s.Hosts = append(s.Hosts, elem)
			}
		}
	}
	{
		{
			s.Images = nil
			for i := 0; i < 0; i++ {
				var elem NilDataDumpImagesItem
				{
					elem.SetFake()
				}
				s.Images = append(s.Images, elem)
			}
		}
	}
	{
		{
			s.Users = nil
			for i := 0; i < 0; i++ {
				var elem DataDumpUsersItem
				{
					elem.SetFake()
				}
				s.Users = append(s.Users, elem)
			}
		}
	}
}

// SetFake set fake values.
func (s *DataDumpHostsItem) SetFake() {
	{
		{
			s.Bonds = nil
			for i := 0; i < 0; i++ {
				var elem NilDataDumpHostsItemBondsItem
				{
					elem.SetFake()
				}
				s.Bonds = append(s.Bonds, elem)
			}
		}
	}
	{
		{
			s.BootImage.SetFake()
		}
	}
	{
		{
			s.Firmware.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Interfaces = nil
			for i := 0; i < 0; i++ {
				var elem NilDataDumpHostsItemInterfacesItem
				{
					elem.SetFake()
				}
				s.Interfaces = append(s.Interfaces, elem)
			}
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Provision.SetFake()
		}
	}
	{
		{
			s.Tags = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Tags = append(s.Tags, elem)
			}
		}
	}
	{
		{
			s.UID.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *DataDumpHostsItemBondsItem) SetFake() {
	{
		{
			s.Bmc.SetFake()
		}
	}
	{
		{
			s.Fqdn.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Ifname.SetFake()
		}
	}
	{
		{
			s.IP.SetFake()
		}
	}
	{
		{
			s.MAC.SetFake()
		}
	}
	{
		{
			s.Mtu.SetFake()
		}
	}
	{
		{
			s.Peers = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Peers = append(s.Peers, elem)
			}
		}
	}
	{
		{
			s.Vlan.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *DataDumpHostsItemInterfacesItem) SetFake() {
	{
		{
			s.Bmc.SetFake()
		}
	}
	{
		{
			s.Fqdn.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Ifname.SetFake()
		}
	}
	{
		{
			s.IP.SetFake()
		}
	}
	{
		{
			s.MAC.SetFake()
		}
	}
	{
		{
			s.Mtu.SetFake()
		}
	}
	{
		{
			s.Vlan.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *DataDumpImagesItem) SetFake() {
	{
		{
			s.Butane.SetFake()
		}
	}
	{
		{
			s.Cmdline.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Initrd = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Initrd = append(s.Initrd, elem)
			}
		}
	}
	{
		{
			s.Kernel.SetFake()
		}
	}
	{
		{
			s.Liveimg.SetFake()
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.ProvisionTemplate.SetFake()
		}
	}
	{
		{
			s.ProvisionTemplates.SetFake()
		}
	}
	{
		{
			s.UID.SetFake()
		}
	}
	{
		{
			s.UserData.SetFake()
		}
	}
	{
		{
			s.Verify.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *DataDumpImagesItemProvisionTemplates) SetFake() {
	var (
		elem NilString
		m    map[string]NilString = s.init()
	)
	for i := 0; i < 0; i++ {
		m[fmt.Sprintf("fake%d", i)] = elem
	}
}

// SetFake set fake values.
func (s *DataDumpUsersItem) SetFake() {
	{
		{
			s.CreatedAt.SetFake()
		}
	}
	{
		{
			s.Hash.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.ModifiedAt.SetFake()
		}
	}
	{
		{
			s.Role.SetFake()
		}
	}
	{
		{
			s.Username.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *Event) SetFake() {
	{
		{
			s.JobMessages = nil
			for i := 0; i < 0; i++ {
				var elem EventJobMessagesItem
				{
					elem.SetFake()
				}
				s.JobMessages = append(s.JobMessages, elem)
			}
		}
	}
	{
		{
			s.Message.SetFake()
		}
	}
	{
		{
			s.Severity.SetFake()
		}
	}
	{
		{
			s.Time.SetFake()
		}
	}
	{
		{
			s.User.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *EventJobMessagesItem) SetFake() {
	{
		{
			s.Host.SetFake()
		}
	}
	{
		{
			s.Msg.SetFake()
		}
	}
	{
		{
			s.RedfishError.SetFake()
		}
	}
	{
		{
			s.Status.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *EventJobMessagesItemRedfishError) SetFake() {
	{
		{
			s.Code.SetFake()
		}
	}
	{
		{
			s.Error.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *EventJobMessagesItemRedfishErrorError) SetFake() {
	{
		{
			s.MessageDotExtendedInfo = nil
			for i := 0; i < 0; i++ {
				var elem EventJobMessagesItemRedfishErrorErrorMessageDotExtendedInfoItem
				{
					elem.SetFake()
				}
				s.MessageDotExtendedInfo = append(s.MessageDotExtendedInfo, elem)
			}
		}
	}
	{
		{
			s.Code.SetFake()
		}
	}
	{
		{
			s.Message.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *EventJobMessagesItemRedfishErrorErrorMessageDotExtendedInfoItem) SetFake() {
	{
		{
			s.Message.SetFake()
		}
	}
	{
		{
			s.MessageArgsDotOdataDotCount.SetFake()
		}
	}
	{
		{
			s.MessageId.SetFake()
		}
	}
	{
		{
			s.RelatedPropertiesDotOdataDotCount.SetFake()
		}
	}
	{
		{
			s.Resolution.SetFake()
		}
	}
	{
		{
			s.Severity.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *GenericResponse) SetFake() {
	{
		{
			s.Changed.SetFake()
		}
	}
	{
		{
			s.Detail.SetFake()
		}
	}
	{
		{
			s.Title.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *HTTPError) SetFake() {
	{
		{
			s.Detail.SetFake()
		}
	}
	{
		{
			s.Errors.SetFake()
		}
	}
	{
		{
			s.Instance.SetFake()
		}
	}
	{
		{
			s.Status.SetFake()
		}
	}
	{
		{
			s.Title.SetFake()
		}
	}
	{
		{
			s.Type.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *HTTPErrorErrorsItem) SetFake() {
	{
		{
			s.More.SetFake()
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Reason.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *HTTPErrorErrorsItemMore) SetFake() {
	var (
		elem jx.Raw
		m    map[string]jx.Raw = s.init()
	)
	for i := 0; i < 0; i++ {
		m[fmt.Sprintf("fake%d", i)] = elem
	}
}

// SetFake set fake values.
func (s *Host) SetFake() {
	{
		{
			s.Bonds = nil
			for i := 0; i < 0; i++ {
				var elem NilHostBondsItem
				{
					elem.SetFake()
				}
				s.Bonds = append(s.Bonds, elem)
			}
		}
	}
	{
		{
			s.BootImage.SetFake()
		}
	}
	{
		{
			s.Firmware.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Interfaces = nil
			for i := 0; i < 0; i++ {
				var elem NilHostInterfacesItem
				{
					elem.SetFake()
				}
				s.Interfaces = append(s.Interfaces, elem)
			}
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Provision.SetFake()
		}
	}
	{
		{
			s.Tags = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Tags = append(s.Tags, elem)
			}
		}
	}
	{
		{
			s.UID.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *HostBondsItem) SetFake() {
	{
		{
			s.Bmc.SetFake()
		}
	}
	{
		{
			s.Fqdn.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Ifname.SetFake()
		}
	}
	{
		{
			s.IP.SetFake()
		}
	}
	{
		{
			s.MAC.SetFake()
		}
	}
	{
		{
			s.Mtu.SetFake()
		}
	}
	{
		{
			s.Peers = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Peers = append(s.Peers, elem)
			}
		}
	}
	{
		{
			s.Vlan.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *HostInterfacesItem) SetFake() {
	{
		{
			s.Bmc.SetFake()
		}
	}
	{
		{
			s.Fqdn.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Ifname.SetFake()
		}
	}
	{
		{
			s.IP.SetFake()
		}
	}
	{
		{
			s.MAC.SetFake()
		}
	}
	{
		{
			s.Mtu.SetFake()
		}
	}
	{
		{
			s.Vlan.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *JobMessage) SetFake() {
	{
		{
			s.Host.SetFake()
		}
	}
	{
		{
			s.Msg.SetFake()
		}
	}
	{
		{
			s.RedfishError.SetFake()
		}
	}
	{
		{
			s.Status.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *JobMessageRedfishError) SetFake() {
	{
		{
			s.Code.SetFake()
		}
	}
	{
		{
			s.Error.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *JobMessageRedfishErrorError) SetFake() {
	{
		{
			s.MessageDotExtendedInfo.SetFake()
		}
	}
	{
		{
			s.Code.SetFake()
		}
	}
	{
		{
			s.Message.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *JobMessageRedfishErrorErrorMessageDotExtendedInfoItem) SetFake() {
	{
		{
			s.Message.SetFake()
		}
	}
	{
		{
			s.MessageArgsDotOdataDotCount.SetFake()
		}
	}
	{
		{
			s.MessageId.SetFake()
		}
	}
	{
		{
			s.RelatedPropertiesDotOdataDotCount.SetFake()
		}
	}
	{
		{
			s.Resolution.SetFake()
		}
	}
	{
		{
			s.Severity.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *NilBootImageAddRequestBootImagesItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilDataDumpHostsItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilDataDumpHostsItemBondsItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilDataDumpHostsItemInterfacesItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilDataDumpImagesItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilHostBondsItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilHostInterfacesItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilInt) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilNodeAddRequestNodeListItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilNodeAddRequestNodeListItemBondsItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilNodeAddRequestNodeListItemInterfacesItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilRedfishJobJobsItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilRedfishMetricReportReportsItem) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NilString) SetFake() {
	s.Null = true
}

// SetFake set fake values.
func (s *NodeAddRequest) SetFake() {
	{
		{
			s.NodeList = nil
			for i := 0; i < 0; i++ {
				var elem NilNodeAddRequestNodeListItem
				{
					elem.SetFake()
				}
				s.NodeList = append(s.NodeList, elem)
			}
		}
	}
}

// SetFake set fake values.
func (s *NodeAddRequestNodeListItem) SetFake() {
	{
		{
			s.Bonds = nil
			for i := 0; i < 0; i++ {
				var elem NilNodeAddRequestNodeListItemBondsItem
				{
					elem.SetFake()
				}
				s.Bonds = append(s.Bonds, elem)
			}
		}
	}
	{
		{
			s.BootImage.SetFake()
		}
	}
	{
		{
			s.Firmware.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Interfaces = nil
			for i := 0; i < 0; i++ {
				var elem NilNodeAddRequestNodeListItemInterfacesItem
				{
					elem.SetFake()
				}
				s.Interfaces = append(s.Interfaces, elem)
			}
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Provision.SetFake()
		}
	}
	{
		{
			s.Tags = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Tags = append(s.Tags, elem)
			}
		}
	}
	{
		{
			s.UID.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *NodeAddRequestNodeListItemBondsItem) SetFake() {
	{
		{
			s.Bmc.SetFake()
		}
	}
	{
		{
			s.Fqdn.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Ifname.SetFake()
		}
	}
	{
		{
			s.IP.SetFake()
		}
	}
	{
		{
			s.MAC.SetFake()
		}
	}
	{
		{
			s.Mtu.SetFake()
		}
	}
	{
		{
			s.Peers = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.Peers = append(s.Peers, elem)
			}
		}
	}
	{
		{
			s.Vlan.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *NodeAddRequestNodeListItemInterfacesItem) SetFake() {
	{
		{
			s.Bmc.SetFake()
		}
	}
	{
		{
			s.Fqdn.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Ifname.SetFake()
		}
	}
	{
		{
			s.IP.SetFake()
		}
	}
	{
		{
			s.MAC.SetFake()
		}
	}
	{
		{
			s.Mtu.SetFake()
		}
	}
	{
		{
			s.Vlan.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *NodeBootImageRequest) SetFake() {
	{
		{
			s.Image.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *NodeBootTokenResponse) SetFake() {
	{
		{
			s.Nodes = nil
			for i := 0; i < 0; i++ {
				var elem NodeBootTokenResponseNodesItem
				{
					elem.SetFake()
				}
				s.Nodes = append(s.Nodes, elem)
			}
		}
	}
}

// SetFake set fake values.
func (s *NodeBootTokenResponseNodesItem) SetFake() {
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Token.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *NodeProvisionRequest) SetFake() {
	{
		{
			s.Provision.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *NodeTagsRequest) SetFake() {
	{
		{
			s.Tags.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *OptBool) SetFake() {
	var elem bool
	{
		elem = true
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptDateTime) SetFake() {
	var elem time.Time
	{
		elem = time.Now()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptEventJobMessagesItemRedfishError) SetFake() {
	var elem EventJobMessagesItemRedfishError
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptEventJobMessagesItemRedfishErrorError) SetFake() {
	var elem EventJobMessagesItemRedfishErrorError
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptFloat32) SetFake() {
	var elem float32
	{
		elem = float32(0)
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptHTTPErrorErrorsItemMore) SetFake() {
	var elem HTTPErrorErrorsItemMore
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptInt) SetFake() {
	var elem int
	{
		elem = int(0)
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptInt64) SetFake() {
	var elem int64
	{
		elem = int64(0)
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptJobMessageRedfishError) SetFake() {
	var elem JobMessageRedfishError
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptJobMessageRedfishErrorError) SetFake() {
	var elem JobMessageRedfishErrorError
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptNilBootImageAddRequestBootImagesItemProvisionTemplates) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilBootImageProvisionTemplates) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilDataDumpImagesItemProvisionTemplates) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilHTTPErrorErrorsItemArray) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilInt) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilInt64) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilJobMessageRedfishErrorErrorMessageDotExtendedInfoItemArray) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilNilIntArray) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilNilStringArray) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptNilString) SetFake() {
	s.Null = true
	s.Set = true
}

// SetFake set fake values.
func (s *OptRedfishJobJobsItemPayload) SetFake() {
	var elem RedfishJobJobsItemPayload
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptRedfishJobJobsItemSchedule) SetFake() {
	var elem RedfishJobJobsItemSchedule
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptRedfishSystemOem) SetFake() {
	var elem RedfishSystemOem
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptRedfishSystemOemDell) SetFake() {
	var elem RedfishSystemOemDell
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptRedfishSystemOemDellDellSystem) SetFake() {
	var elem RedfishSystemOemDellDellSystem
	{
		elem.SetFake()
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *OptString) SetFake() {
	var elem string
	{
		elem = "string"
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *RedfishJob) SetFake() {
	{
		{
			s.Jobs = nil
			for i := 0; i < 0; i++ {
				var elem NilRedfishJobJobsItem
				{
					elem.SetFake()
				}
				s.Jobs = append(s.Jobs, elem)
			}
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishJobJobsItem) SetFake() {
	{
		{
			s.OdataDotContext.SetFake()
		}
	}
	{
		{
			s.OdataDotID.SetFake()
		}
	}
	{
		{
			s.OdataDotType.SetFake()
		}
	}
	{
		{
			s.CreatedBy.SetFake()
		}
	}
	{
		{
			s.Description.SetFake()
		}
	}
	{
		{
			s.EndTime.SetFake()
		}
	}
	{
		{
			s.EstimatedDuration.SetFake()
		}
	}
	{
		{
			s.HidePayload.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.JobState.SetFake()
		}
	}
	{
		{
			s.JobStatus.SetFake()
		}
	}
	{
		{
			s.MaxExecutionTime.SetFake()
		}
	}
	{
		{
			s.Messages = nil
			for i := 0; i < 0; i++ {
				var elem RedfishJobJobsItemMessagesItem
				{
					elem.SetFake()
				}
				s.Messages = append(s.Messages, elem)
			}
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Payload.SetFake()
		}
	}
	{
		{
			s.PercentComplete.SetFake()
		}
	}
	{
		{
			s.Schedule.SetFake()
		}
	}
	{
		{
			s.StartTime.SetFake()
		}
	}
	{
		{
			s.StepOrder.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishJobJobsItemMessagesItem) SetFake() {
	{
		{
			s.OdataDotID.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.Message.SetFake()
		}
	}
	{
		{
			s.MessageArgs = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.MessageArgs = append(s.MessageArgs, elem)
			}
		}
	}
	{
		{
			s.MessageId.SetFake()
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.RelatedProperties.SetFake()
		}
	}
	{
		{
			s.Resolution.SetFake()
		}
	}
	{
		{
			s.Severity.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishJobJobsItemPayload) SetFake() {
	{
		{
			s.HttpHeaders.SetFake()
		}
	}
	{
		{
			s.HttpOperation.SetFake()
		}
	}
	{
		{
			s.JsonBody.SetFake()
		}
	}
	{
		{
			s.TargetUri.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishJobJobsItemSchedule) SetFake() {
	{
		{
			s.EnabledDaysOfMonth.SetFake()
		}
	}
	{
		{
			s.EnabledDaysOfWeek.SetFake()
		}
	}
	{
		{
			s.EnabledIntervals.SetFake()
		}
	}
	{
		{
			s.EnabledMonthsOfYear.SetFake()
		}
	}
	{
		{
			s.InitialStartTime.SetFake()
		}
	}
	{
		{
			s.Lifetime.SetFake()
		}
	}
	{
		{
			s.MaxOccurrences.SetFake()
		}
	}
	{
		{
			s.RecurrenceInterval.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishMetricReport) SetFake() {
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Reports = nil
			for i := 0; i < 0; i++ {
				var elem NilRedfishMetricReportReportsItem
				{
					elem.SetFake()
				}
				s.Reports = append(s.Reports, elem)
			}
		}
	}
}

// SetFake set fake values.
func (s *RedfishMetricReportReportsItem) SetFake() {
	{
		{
			s.OdataDotContext.SetFake()
		}
	}
	{
		{
			s.OdataDotEtag.SetFake()
		}
	}
	{
		{
			s.OdataDotID.SetFake()
		}
	}
	{
		{
			s.OdataDotType.SetFake()
		}
	}
	{
		{
			s.Context.SetFake()
		}
	}
	{
		{
			s.Description.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.MetricValues = nil
			for i := 0; i < 0; i++ {
				var elem RedfishMetricReportReportsItemMetricValuesItem
				{
					elem.SetFake()
				}
				s.MetricValues = append(s.MetricValues, elem)
			}
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Oem = []byte("null")
		}
	}
	{
		{
			s.Timestamp.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishMetricReportReportsItemMetricValuesItem) SetFake() {
	{
		{
			s.MetricID.SetFake()
		}
	}
	{
		{
			s.MetricProperty.SetFake()
		}
	}
	{
		{
			s.MetricValue.SetFake()
		}
	}
	{
		{
			s.Oem = []byte("null")
		}
	}
	{
		{
			s.Timestamp.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishSystem) SetFake() {
	{
		{
			s.BiosVersion.SetFake()
		}
	}
	{
		{
			s.BootNext.SetFake()
		}
	}
	{
		{
			s.BootOrder = nil
			for i := 0; i < 0; i++ {
				var elem string
				{
					elem = "string"
				}
				s.BootOrder = append(s.BootOrder, elem)
			}
		}
	}
	{
		{
			s.Health.SetFake()
		}
	}
	{
		{
			s.HostName.SetFake()
		}
	}
	{
		{
			s.Manufacturer.SetFake()
		}
	}
	{
		{
			s.Model.SetFake()
		}
	}
	{
		{
			s.Name.SetFake()
		}
	}
	{
		{
			s.Oem.SetFake()
		}
	}
	{
		{
			s.PowerStatus.SetFake()
		}
	}
	{
		{
			s.ProcessorCount.SetFake()
		}
	}
	{
		{
			s.SerialNumber.SetFake()
		}
	}
	{
		{
			s.TotalMemory.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishSystemOem) SetFake() {
	{
		{
			s.Dell.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishSystemOemDell) SetFake() {
	{
		{
			s.DellSystem.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *RedfishSystemOemDellDellSystem) SetFake() {
	{
		{
			s.ManagedSystemSize.SetFake()
		}
	}
	{
		{
			s.MaxCPUSockets.SetFake()
		}
	}
	{
		{
			s.MaxDIMMSlots.SetFake()
		}
	}
	{
		{
			s.MaxPCIeSlots.SetFake()
		}
	}
	{
		{
			s.SystemID.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *User) SetFake() {
	{
		{
			s.CreatedAt.SetFake()
		}
	}
	{
		{
			s.Hash.SetFake()
		}
	}
	{
		{
			s.ID.SetFake()
		}
	}
	{
		{
			s.ModifiedAt.SetFake()
		}
	}
	{
		{
			s.Role.SetFake()
		}
	}
	{
		{
			s.Username.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *UserRoleRequest) SetFake() {
	{
		{
			s.Role.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *UserStoreRequest) SetFake() {
	{
		{
			s.Password.SetFake()
		}
	}
	{
		{
			s.Username.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *UserStoreResponse) SetFake() {
	{
		{
			s.Role.SetFake()
		}
	}
	{
		{
			s.Username.SetFake()
		}
	}
}
