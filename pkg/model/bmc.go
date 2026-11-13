// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"github.com/stmcginnis/gofish/oem/dell"
)

// JobMessage is one node's outcome as the event log records it
type JobMessage struct {
	Status string `json:"status"`
	Host   string `json:"host"`
	Msg    string `json:"msg"`
}

type JobMessageList []JobMessage

type RedfishSystem struct {
	Name           string  `json:"name"`
	HostName       string  `json:"host_name"`
	BIOSVersion    string  `json:"bios_version"`
	SerialNumber   string  `json:"serial_number"`
	Manufacturer   string  `json:"manufacturer"`
	Model          string  `json:"model"`
	PowerStatus    string  `json:"power_status"`
	Health         string  `json:"health"`
	TotalMemory    float32 `json:"total_memory"`
	ProcessorCount int     `json:"processor_count"`
	ProcessorModel string  `json:"processor_model"`
	// ProcessorSockets and ProcessorCores are zero when the bmc does not report them
	ProcessorSockets int `json:"processor_sockets"`
	ProcessorCores   int `json:"processor_cores"`
	// BootProgress is the last boot state the bmc saw, such as OSRunning, and BootProgressTime when it got there. Empty on firmware that does not report boot progress
	BootProgress       string `json:"boot_progress"`
	BootProgressTime   string `json:"boot_progress_time"`
	BootOverrideTarget string `json:"boot_override_target"`
	// BootOverrideEnabled is Disabled, Once or Continuous
	BootOverrideEnabled string          `json:"boot_override_enabled"`
	BootMode            string          `json:"boot_mode" description:"UEFI or Legacy"`
	BootNext            string          `json:"boot_next"`
	BootOrder           []string        `json:"boot_order" oai3:"nullable"`
	LocationIndicator   string          `json:"location_indicator" description:"On, Off, or empty when the bmc does not report one"`
	AssetTag            string          `json:"asset_tag"`
	Location            *BmcLocation    `json:"location" oai3:"nullable"`
	OEMDell             *dell.OEMSystem `json:"oem_dell" oai3:"nullable"`
}

// BmcSummary counts a run. FailedNodes is a nodeset string, so it can be pasted straight back into another command
type BmcSummary struct {
	Total       int    `json:"total"`
	Succeeded   int    `json:"succeeded"`
	Failed      int    `json:"failed"`
	FailedNodes string `json:"failed_nodes"`
	DurationMs  int64  `json:"duration_ms"`
}

// BmcResult is one node's outcome. Message says why a node failed. Data is present only on success, and only when the operation returns anything
type BmcResult[T any] struct {
	Node       string `json:"node"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	DurationMs int64  `json:"duration_ms"`
	Data       *T     `json:"data" oai3:"nullable"`
}

// BmcResponse is the envelope of every bmc route. RunID appears in the server log and the event record, which ties them to this response
type BmcResponse[T any] struct {
	RunID   string         `json:"run_id"`
	Summary BmcSummary     `json:"summary"`
	Results []BmcResult[T] `json:"results" oai3:"nullable"`
}

// BmcAck is the payload of an operation that reports nothing but whether it worked, so data is always null. What a node has to say, such as a power reset being substituted, arrives in the result's message
type BmcAck struct{}

// BmcJob is one job queued on a BMC. It is Grendel's own shape rather than gofish's schemas.Job, whose Messages hold slices with no omitempty. Those serialize as null, which the generated client rejects
type BmcJob struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	State           string `json:"state"`
	Status          string `json:"status"`
	Type            string `json:"type"`
	PercentComplete int    `json:"percent_complete"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Message         string `json:"message"`
}

// BmcJobList is the payload of a jobs query. It is a struct rather than a bare slice so a node with an empty queue is distinguishable from one that was never reached
type BmcJobList struct {
	Jobs []BmcJob `json:"jobs" oai3:"nullable"`
}

// BmcSELEntry is one system event log record, narrowed the same way BmcJob is. gofish's schemas.LogEntry carries the diagnostic data blob and several slices with no omitempty
type BmcSELEntry struct {
	ID         string `json:"id"`
	Created    string `json:"created"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	MessageID  string `json:"message_id"`
	EntryType  string `json:"entry_type"`
	SensorType string `json:"sensor_type"`
	Resolution string `json:"resolution"`
}

// BmcSELLog is one node's system event log, newest entry first
type BmcSELLog struct {
	Entries []BmcSELEntry `json:"entries" oai3:"nullable"`
}

// BmcFirmwareUpdate is one package the iDRAC found newer in the catalog than what is installed
type BmcFirmwareUpdate struct {
	Name             string `json:"name"`
	DisplayName      string `json:"display_name"`
	ComponentType    string `json:"component_type"`
	InstalledVersion string `json:"installed_version"`
	PackageVersion   string `json:"package_version"`
	Criticality      string `json:"criticality" description:"1 = Warning, 2 = Critical, 3 = Recommended"`
	RebootType       string `json:"reboot_type" description:"NONE, IDRAC (non disruptive) or HOST (full power cycle)"`
	PackagePath      string `json:"package_path"`
}

// BmcFirmwareUpdates is one node's available updates. RebootType is the most disruptive reboot any of them needs
type BmcFirmwareUpdates struct {
	Updates    []BmcFirmwareUpdate `json:"updates" oai3:"nullable"`
	RebootType string              `json:"reboot_type" description:"NONE, IDRAC or HOST"`
}

// BmcBios is one node's BIOS attributes. Values keep the JSON type the BMC reports, a string, number or boolean. Pending holds the staged values that differ from the applied ones, which take effect at the apply time they were staged with
type BmcBios struct {
	Attributes        map[string]any `json:"attributes"`
	Pending           map[string]any `json:"pending"`
	AttributeRegistry string         `json:"attribute_registry" description:"Name of the attribute registry describing these attributes, empty when the bmc names none"`
	ApplyTimes        []string       `json:"apply_times" description:"Apply times the bmc accepts for a change"`
	ResetPending      bool           `json:"reset_pending" description:"A reset to defaults is staged"`
	Model             string         `json:"model"`
	BIOSVersion       string         `json:"bios_version"`
}

// BmcBiosValue is one allowed value of an enumeration attribute
type BmcBiosValue struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// BmcBiosAttribute is one attribute registry entry, narrowed to what an editor needs. Bounds and lengths are null when the registry sets none
type BmcBiosAttribute struct {
	Name          string         `json:"name"`
	DisplayName   string         `json:"display_name"`
	HelpText      string         `json:"help_text"`
	WarningText   string         `json:"warning_text"`
	MenuPath      string         `json:"menu_path"`
	DisplayOrder  int            `json:"display_order"`
	Type          string         `json:"type" description:"Enumeration, String, Integer, Boolean or Password"`
	Values        []BmcBiosValue `json:"values"`
	DefaultValue  any            `json:"default_value" oai3:"nullable"`
	ReadOnly      bool           `json:"read_only"`
	WriteOnly     bool           `json:"write_only"`
	Hidden        bool           `json:"hidden"`
	Immutable     bool           `json:"immutable"`
	GrayOut       bool           `json:"gray_out"`
	ResetRequired bool           `json:"reset_required"`
	SystemUnique  bool           `json:"system_unique" description:"The value is meant to differ per system, such as an asset tag"`
	LowerBound    *int64         `json:"lower_bound" oai3:"nullable"`
	UpperBound    *int64         `json:"upper_bound" oai3:"nullable"`
	MinLength     *int           `json:"min_length" oai3:"nullable"`
	MaxLength     *int           `json:"max_length" oai3:"nullable"`
}

// BmcBiosRegistry describes a node's BIOS attributes. Available is false when the bmc publishes no registry it serves itself, and Attributes is then empty
type BmcBiosRegistry struct {
	Available  bool               `json:"available"`
	Registry   string             `json:"registry"`
	Version    string             `json:"version"`
	Language   string             `json:"language"`
	Attributes []BmcBiosAttribute `json:"attributes"`
}

// BmcSecureBoot is one node's UEFI secure boot state, with the keys when asked for. Enabled is what the next boot uses, and PendingEnabled is a staged change to it, nil when none is staged
type BmcSecureBoot struct {
	Enabled        bool     `json:"enabled"`
	PendingEnabled *bool    `json:"pending_enabled" oai3:"nullable"`
	Mode           string   `json:"mode" description:"SetupMode, UserMode, AuditMode or DeployedMode"`
	CurrentBoot    string   `json:"current_boot" description:"Whether secure boot was in effect for the current boot: Enabled or Disabled"`
	Databases      []string `json:"databases" description:"The key databases the bmc exposes, such as PK, KEK, db and dbx"`
	// Keys is the contents of the databases, present only when they were asked for
	Keys []BmcSecureBootDatabase `json:"keys,omitempty"`
}

// BmcSecureBootCert is one certificate in a secure boot database. Fingerprint is the SHA-256 of the certificate, empty when the bmc reports neither the certificate nor a SHA-256 fingerprint
type BmcSecureBootCert struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	Issuer      string `json:"issuer"`
	NotBefore   string `json:"not_before"`
	NotAfter    string `json:"not_after"`
	Fingerprint string `json:"fingerprint"`
	Owner       string `json:"owner" description:"UEFI signature owner GUID"`
}

// BmcSecureBootSignature is one hash in a secure boot database, such as a revoked binary in dbx
type BmcSecureBootSignature struct {
	ID    string `json:"id"`
	Type  string `json:"type" description:"UEFI signature type, such as EFI_CERT_SHA256_GUID"`
	Hash  string `json:"hash" description:"Hex encoded signature data"`
	Owner string `json:"owner" description:"UEFI signature owner GUID"`
}

// BmcSecureBootDatabase is the contents of one secure boot key database
type BmcSecureBootDatabase struct {
	ID           string                   `json:"id"`
	Certificates []BmcSecureBootCert      `json:"certificates"`
	Signatures   []BmcSecureBootSignature `json:"signatures"`
}

// BmcLocation is where a node sits, as its bmc records it. RackUnit is nil when unset
type BmcLocation struct {
	Datacenter string `json:"datacenter"`
	Room       string `json:"room"`
	Row        string `json:"row"`
	Rack       string `json:"rack"`
	RackUnit   *int   `json:"rack_unit" oai3:"nullable"`
}

// BmcAssetTagResult is one node's asset tag before and after a change. Changed is false when the bmc already had it, and on a dry run nothing was written either way
type BmcAssetTagResult struct {
	Serial  string `json:"serial"`
	Current string `json:"current"`
	Desired string `json:"desired"`
	Changed bool   `json:"changed"`
}

// BmcLocationResult is one node's location before and after a change
type BmcLocationResult struct {
	Current BmcLocation `json:"current"`
	Desired BmcLocation `json:"desired"`
	Changed bool        `json:"changed"`
}
