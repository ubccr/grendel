package model

import "github.com/stmcginnis/gofish/redfish"

type OSPower struct {
	Hosts       string                           `json:"hosts"`
	PowerOption redfish.ResetType                `json:"power_option"`
	BootOption  redfish.BootSourceOverrideTarget `json:"boot_option"`
}

type JobMessageList []JobMessage

type JobMessage struct {
	Status       string       `json:"status"`
	Host         string       `json:"host"`
	Msg          string       `json:"msg"`
	RedfishError RedfishError `json:"redfish_error"`
}

type RedfishJobList []RedfishJob
type RedfishJob struct {
	Host string         `json:"name"`
	Jobs []*redfish.Job `json:"jobs"`
}

type RedfishMetricReportList []RedfishMetricReport
type RedfishMetricReport struct {
	Name    string                  `json:"name"`
	Reports []*redfish.MetricReport `json:"reports"`
}

type RedfishSystemList []RedfishSystem
type RedfishSystem struct {
	Name           string           `json:"name"`
	HostName       string           `json:"host_name"`
	BIOSVersion    string           `json:"bios_version"`
	SerialNumber   string           `json:"serial_number"`
	Manufacturer   string           `json:"manufacturer"`
	Model          string           `json:"model"`
	PowerStatus    string           `json:"power_status"`
	Health         string           `json:"health"`
	TotalMemory    float32          `json:"total_memory"`
	ProcessorCount int              `json:"processor_count"`
	BootNext       string           `json:"boot_next"`
	BootOrder      []string         `json:"boot_order"`
	OEM            RedfishSystemOEM `json:"oem"`
}

type RedfishSystemOEM struct {
	Dell struct {
		DellSystem struct {
			ManagedSystemSize string `json:"ManagedSystemSize"`
			MaxCPUSockets     int    `json:"MaxCPUSockets"`
			MaxDIMMSlots      int    `json:"MaxDIMMSlots"`
			MaxPCIeSlots      int    `json:"MaxPCIeSlots"`
			SystemID          int    `json:"SystemID"`
		} `json:"DellSystem"`
	} `json:"Dell"`
}

// TODO: verify correct json parsing
type RedfishError struct {
	Code  string `json:"code"`
	Error struct {
		MessageExtendedInfo []struct {
			Message                string `json:"Message"`
			MessageArgsCount       int    `json:"MessageArgs.@odata.count"`
			MessageId              string `json:"MessageId"`
			RelatedPropertiesCount int    `json:"RelatedProperties.@odata.count"`
			Resolution             string `json:"Resolution"`
			Severity               string `json:"Severity"`
		} `json:"@Message.ExtendedInfo,omitempty"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
