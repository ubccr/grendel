// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"github.com/stmcginnis/gofish/oem/dell"
)

type RedfishSystem struct {
	Name           string  `json:"name" validate:"required"`
	HostName       string  `json:"host_name" validate:"required"`
	BIOSVersion    string  `json:"bios_version" validate:"required"`
	SerialNumber   string  `json:"serial_number" validate:"required"`
	Manufacturer   string  `json:"manufacturer" validate:"required"`
	Model          string  `json:"model" validate:"required"`
	PowerStatus    string  `json:"power_status" validate:"required"`
	Health         string  `json:"health" validate:"required"`
	TotalMemory    float32 `json:"total_memory" validate:"required"`
	ProcessorCount int     `json:"processor_count" validate:"required"`
	ProcessorModel string  `json:"processor_model" validate:"required"`
	// ProcessorSockets and ProcessorCores are zero when the bmc does not report them
	ProcessorSockets int `json:"processor_sockets" validate:"required"`
	ProcessorCores   int `json:"processor_cores" validate:"required"`
	// BootProgress is the last boot state the bmc saw, such as OSRunning, and BootProgressTime when it got there. Empty on firmware that does not report boot progress
	BootProgress       string `json:"boot_progress" validate:"required"`
	BootProgressTime   string `json:"boot_progress_time" validate:"required"`
	BootOverrideTarget string `json:"boot_override_target" validate:"required"`
	// BootOverrideEnabled is Disabled, Once or Continuous
	BootOverrideEnabled string          `json:"boot_override_enabled" validate:"required"`
	BootMode            string          `json:"boot_mode" validate:"required" description:"UEFI or Legacy"`
	BootNext            string          `json:"boot_next" validate:"required"`
	BootOrder           []string        `json:"boot_order" oai3:"nullable"`
	LocationIndicator   string          `json:"location_indicator" validate:"required" description:"On, Off, or empty when the bmc does not report one"`
	AssetTag            string          `json:"asset_tag" validate:"required"`
	Location            *BmcLocation    `json:"location" oai3:"nullable"`
	OEMDell             *dell.OEMSystem `json:"oem_dell" oai3:"nullable"`
}

// BmcSummary counts a run. FailedNodes is a nodeset string, so it can be pasted straight back into another command
type BmcSummary struct {
	Total       int    `json:"total" validate:"required"`
	Succeeded   int    `json:"succeeded" validate:"required"`
	Failed      int    `json:"failed" validate:"required"`
	FailedNodes string `json:"failed_nodes" validate:"required"`
	DurationMs  int64  `json:"duration_ms" validate:"required"`
}

// BmcResult is one node's outcome. Message says why a node failed. Data is present only on success, and only when the operation returns anything
type BmcResult[T any] struct {
	Node       string `json:"node" validate:"required"`
	Success    bool   `json:"success" validate:"required"`
	Message    string `json:"message" validate:"required"`
	DurationMs int64  `json:"duration_ms" validate:"required"`
	Data       *T     `json:"data" oai3:"nullable"`
}

// BmcResponse is the envelope of every bmc route. RunID appears in the server log and the event record, which ties them to this response. Its types are never request bodies, so validate:"required" validates nothing and only marks a field as always sent, which keeps it from being optional in the generated clients
type BmcResponse[T any] struct {
	RunID   string         `json:"run_id" validate:"required"`
	Summary BmcSummary     `json:"summary" validate:"required"`
	Results []BmcResult[T] `json:"results" validate:"required"`
}

// BmcAck is the payload of an operation that reports nothing but whether it worked, so data is always null. What a node has to say, such as a power reset being substituted, arrives in the result's message
type BmcAck struct{}

// BmcJob is one job queued on a BMC. It is Grendel's own shape rather than gofish's schemas.Job, whose Messages hold slices with no omitempty. Those serialize as null, which the generated client rejects
type BmcJob struct {
	ID              string `json:"id" validate:"required"`
	Name            string `json:"name" validate:"required"`
	State           string `json:"state" validate:"required"`
	Status          string `json:"status" validate:"required"`
	Type            string `json:"type" validate:"required"`
	PercentComplete int    `json:"percent_complete" validate:"required"`
	StartTime       string `json:"start_time" validate:"required"`
	EndTime         string `json:"end_time" validate:"required"`
	Message         string `json:"message" validate:"required"`
}

// BmcJobList is the payload of a jobs query. It is a struct rather than a bare slice so a node with an empty queue is distinguishable from one that was never reached
type BmcJobList struct {
	Jobs []BmcJob `json:"jobs" oai3:"nullable"`
}

// BmcSELEntry is one system event log record, narrowed the same way BmcJob is. gofish's schemas.LogEntry carries the diagnostic data blob and several slices with no omitempty
type BmcSELEntry struct {
	ID         string `json:"id" validate:"required"`
	Created    string `json:"created" validate:"required"`
	Severity   string `json:"severity" validate:"required"`
	Message    string `json:"message" validate:"required"`
	MessageID  string `json:"message_id" validate:"required"`
	EntryType  string `json:"entry_type" validate:"required"`
	SensorType string `json:"sensor_type" validate:"required"`
	Resolution string `json:"resolution" validate:"required"`
}

// BmcSELLog is one node's system event log, newest entry first
type BmcSELLog struct {
	Entries []BmcSELEntry `json:"entries" oai3:"nullable"`
}

// BmcFirmwareUpdate is one package the iDRAC found newer in the catalog than what is installed
type BmcFirmwareUpdate struct {
	Name             string `json:"name" validate:"required"`
	DisplayName      string `json:"display_name" validate:"required"`
	ComponentType    string `json:"component_type" validate:"required"`
	InstalledVersion string `json:"installed_version" validate:"required"`
	PackageVersion   string `json:"package_version" validate:"required"`
	Criticality      string `json:"criticality" validate:"required" description:"1 = Warning, 2 = Critical, 3 = Recommended"`
	RebootType       string `json:"reboot_type" validate:"required" description:"NONE, IDRAC (non disruptive) or HOST (full power cycle)"`
	PackagePath      string `json:"package_path" validate:"required"`
}

// BmcFirmwareUpdates is one node's available updates. RebootType is the most disruptive reboot any of them needs
type BmcFirmwareUpdates struct {
	Updates    []BmcFirmwareUpdate `json:"updates" oai3:"nullable"`
	RebootType string              `json:"reboot_type" validate:"required" description:"NONE, IDRAC or HOST"`
}

// BmcBios is one node's BIOS attributes. Values keep the JSON type the BMC reports, a string, number or boolean. Pending holds the staged values that differ from the applied ones, which take effect at the apply time they were staged with
type BmcBios struct {
	Attributes        map[string]any `json:"attributes"`
	Pending           map[string]any `json:"pending"`
	AttributeRegistry string         `json:"attribute_registry" validate:"required" description:"Name of the attribute registry describing these attributes, empty when the bmc names none"`
	ApplyTimes        []string       `json:"apply_times" description:"Apply times the bmc accepts for a change"`
	ResetPending      bool           `json:"reset_pending" validate:"required" description:"A reset to defaults is staged"`
	Model             string         `json:"model" validate:"required"`
	BIOSVersion       string         `json:"bios_version" validate:"required"`
}

type BmcBiosValue struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
}

// BmcBiosAttribute is one attribute registry entry, narrowed to what an editor needs. Bounds and lengths are null when the registry sets none
type BmcBiosAttribute struct {
	Name          string         `json:"name" validate:"required"`
	DisplayName   string         `json:"display_name" validate:"required"`
	HelpText      string         `json:"help_text" validate:"required"`
	WarningText   string         `json:"warning_text" validate:"required"`
	MenuPath      string         `json:"menu_path" validate:"required"`
	DisplayOrder  int            `json:"display_order" validate:"required"`
	Type          string         `json:"type" validate:"required" description:"Enumeration, String, Integer, Boolean or Password"`
	Values        []BmcBiosValue `json:"values"`
	DefaultValue  any            `json:"default_value" oai3:"nullable"`
	ReadOnly      bool           `json:"read_only" validate:"required"`
	WriteOnly     bool           `json:"write_only" validate:"required"`
	Hidden        bool           `json:"hidden" validate:"required"`
	Immutable     bool           `json:"immutable" validate:"required"`
	GrayOut       bool           `json:"gray_out" validate:"required"`
	ResetRequired bool           `json:"reset_required" validate:"required"`
	SystemUnique  bool           `json:"system_unique" validate:"required" description:"The value is meant to differ per system, such as an asset tag"`
	LowerBound    *int64         `json:"lower_bound" oai3:"nullable"`
	UpperBound    *int64         `json:"upper_bound" oai3:"nullable"`
	MinLength     *int           `json:"min_length" oai3:"nullable"`
	MaxLength     *int           `json:"max_length" oai3:"nullable"`
}

// BmcBiosRegistry describes a node's BIOS attributes. Available is false when the bmc publishes no registry it serves itself, and Attributes is then empty
type BmcBiosRegistry struct {
	Available  bool               `json:"available" validate:"required"`
	Registry   string             `json:"registry" validate:"required"`
	Version    string             `json:"version" validate:"required"`
	Language   string             `json:"language" validate:"required"`
	Attributes []BmcBiosAttribute `json:"attributes"`
}

// BmcSecureBoot is one node's UEFI secure boot state, with the keys when asked for. Enabled is what the next boot uses, and PendingEnabled is a staged change to it, nil when none is staged
type BmcSecureBoot struct {
	Enabled        bool     `json:"enabled" validate:"required"`
	PendingEnabled *bool    `json:"pending_enabled" oai3:"nullable"`
	Mode           string   `json:"mode" validate:"required" description:"SetupMode, UserMode, AuditMode or DeployedMode"`
	CurrentBoot    string   `json:"current_boot" validate:"required" description:"Whether secure boot was in effect for the current boot: Enabled or Disabled"`
	Databases      []string `json:"databases" description:"The key databases the bmc exposes, such as PK, KEK, db and dbx"`
	// Keys is the contents of the databases, present only when they were asked for
	Keys []BmcSecureBootDatabase `json:"keys,omitempty"`
}

// BmcSecureBootCert is one certificate in a secure boot database. Fingerprint is the SHA-256 of the certificate, empty when the bmc reports neither the certificate nor a SHA-256 fingerprint
type BmcSecureBootCert struct {
	ID          string `json:"id" validate:"required"`
	Subject     string `json:"subject" validate:"required"`
	Issuer      string `json:"issuer" validate:"required"`
	NotBefore   string `json:"not_before" validate:"required"`
	NotAfter    string `json:"not_after" validate:"required"`
	Fingerprint string `json:"fingerprint" validate:"required"`
	Owner       string `json:"owner" validate:"required" description:"UEFI signature owner GUID"`
}

// BmcSecureBootSignature is one hash in a secure boot database, such as a revoked binary in dbx
type BmcSecureBootSignature struct {
	ID    string `json:"id" validate:"required"`
	Type  string `json:"type" validate:"required" description:"UEFI signature type, such as EFI_CERT_SHA256_GUID"`
	Hash  string `json:"hash" validate:"required" description:"Hex encoded signature data"`
	Owner string `json:"owner" validate:"required" description:"UEFI signature owner GUID"`
}

type BmcSecureBootDatabase struct {
	ID           string                   `json:"id" validate:"required"`
	Certificates []BmcSecureBootCert      `json:"certificates"`
	Signatures   []BmcSecureBootSignature `json:"signatures"`
}

// BmcLocation is where a node sits, as its bmc records it. RackUnit is nil when unset
type BmcLocation struct {
	Datacenter string `json:"datacenter" validate:"required"`
	Room       string `json:"room" validate:"required"`
	Row        string `json:"row" validate:"required"`
	Rack       string `json:"rack" validate:"required"`
	RackUnit   *int   `json:"rack_unit" oai3:"nullable"`
}

// BmcAssetTagResult is one node's asset tag before and after a change. Changed is false when the bmc already had it, and on a dry run nothing was written either way
type BmcAssetTagResult struct {
	Serial  string `json:"serial" validate:"required"`
	Current string `json:"current" validate:"required"`
	Desired string `json:"desired" validate:"required"`
	Changed bool   `json:"changed" validate:"required"`
}

type BmcLocationResult struct {
	Current BmcLocation `json:"current" validate:"required"`
	Desired BmcLocation `json:"desired" validate:"required"`
	Changed bool        `json:"changed" validate:"required"`
}

type BmcNIC struct {
	ID        string `json:"id" validate:"required" example:"NIC.Slot.3-1-1"`
	Name      string `json:"name" validate:"required"`
	MAC       string `json:"mac" validate:"required" example:"b8:ca:3a:6e:00:10"`
	Link      string `json:"link" validate:"required" description:"LinkUp, LinkDown or NoLink, or empty when the BMC does not say" example:"LinkUp"`
	SpeedMbps int    `json:"speed_mbps" validate:"required"`
	Enabled   bool   `json:"enabled" validate:"required"`
}

// BmcNICList is the payload of a NIC query. Like BmcJobList it is a struct so a host with no ports reported is distinguishable from one never reached
type BmcNICList struct {
	NICs []BmcNIC `json:"nics" oai3:"nullable"`
}
