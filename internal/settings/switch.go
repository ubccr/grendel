// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import "time"

// GNMIOptions are the gNMI settings, which exist once for every switch and once more per NOS
type GNMIOptions struct {
	Port            Opt[int]
	Plaintext       Opt[bool]
	SkipVerify      Opt[bool]
	CACert          Opt[string]
	NetworkInstance Opt[string]
	Target          Opt[string]
}

var (
	SwitchGNMI       = gnmiOptions("switch.gnmi", "gNMI", "")
	SwitchGNMIArista = gnmiOptions("switch.gnmi.arista", "gNMI on Arista", "switch.gnmi")
	SwitchGNMISonic  = gnmiOptions("switch.gnmi.sonic", "gNMI on SONiC", "switch.gnmi")
)

// gnmiOptions registers the gNMI settings under prefix. With inherit set, each follows the setting of the same name under it until it is set itself.
func gnmiOptions(prefix, group, inherit string) GNMIOptions {
	opt := func(name, summary, help string, kind Kind, def any) Option {
		o := Option{Key: prefix + "." + name, Section: "switch", Group: group, Summary: summary, Help: help, Kind: kind, Default: def, Apply: ApplyLive}
		if inherit != "" {
			o.InheritFrom = inherit + "." + name
		}
		return o
	}
	port := opt("port", "Port", "TCP port of the switch's gNMI server.", KindInt, 6030)
	port.Min, port.Max = ptr(1), ptr(65535)

	return GNMIOptions{
		Port:            register[int](port),
		Plaintext:       register[bool](opt("plaintext", "Plaintext", "Use gRPC with no TLS, which is the EOS default for \"management api gnmi\".", KindBool, false)),
		SkipVerify:      register[bool](opt("skip_verify", "Accept untrusted certificates", "Accept the switch's self-signed certificate. Ignored when a CA certificate is set.", KindBool, true)),
		CACert:          register[string](opt("ca_cert", "CA certificate", "Path on the server to a CA certificate the switch's certificate is checked against.", KindPath, "")),
		NetworkInstance: register[string](opt("network_instance", "Network instance", "Network instance holding the MAC table. Some SONiC builds need * or a VLAN name.", KindString, "default")),
		Target:          register[string](opt("target", "Target", "gNMI target sent in the request prefix, for switches or proxies that serve several. Empty sends none.", KindString, "")),
	}
}

var (
	SwitchUsername = register[string](Option{
		Key:     "switch.username",
		Section: "switch",
		Summary: "Username",
		Help:    "gNMI login used on every switch.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
	})
	SwitchPassword = register[string](Option{
		Key:     "switch.password",
		Section: "switch",
		Summary: "Password",
		Help:    "Password for switch.username.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
		Secret:  true,
	})
	SwitchTimeout = register[time.Duration](Option{
		Key:     "switch.timeout",
		Section: "switch",
		Summary: "Timeout",
		Help:    "How long reading one switch may take.",
		Kind:    KindDuration,
		Default: 30 * time.Second,
		Apply:   ApplyLive,
	})

	SwitchSNMPVersion = register[string](Option{
		Key:     "switch.snmp.version",
		Section: "switch",
		Group:   "SNMP",
		Summary: "Version",
		Help:    "SNMP version spoken to Dell OS10 switches.",
		Kind:    KindEnum,
		Enum:    []string{"2c", "3"},
		Default: "2c",
		Apply:   ApplyLive,
	})
	SwitchSNMPCommunity = register[string](Option{
		Key:     "switch.snmp.community",
		Section: "switch",
		Group:   "SNMP",
		Summary: "Community",
		Help:    "Community string for SNMP 2c. Defaults to public.",
		Kind:    KindString,
		Default: "public",
		Apply:   ApplyLive,
		Secret:  true,
	})
	SwitchSNMPPort = register[int](Option{
		Key:     "switch.snmp.port",
		Section: "switch",
		Group:   "SNMP",
		Summary: "Port",
		Help:    "UDP port of the switch's SNMP agent.",
		Kind:    KindInt,
		Default: 161,
		Min:     ptr(1),
		Max:     ptr(65535),
		Apply:   ApplyLive,
	})
	SwitchSNMPTimeout = register[time.Duration](Option{
		Key:     "switch.snmp.timeout",
		Section: "switch",
		Group:   "SNMP",
		Summary: "Timeout",
		Help:    "How long one SNMP request may take.",
		Kind:    KindDuration,
		Default: 5 * time.Second,
		Apply:   ApplyLive,
	})
	SwitchSNMPRetries = register[int](Option{
		Key:     "switch.snmp.retries",
		Section: "switch",
		Group:   "SNMP",
		Summary: "Retries",
		Help:    "How many times an unanswered SNMP request is sent again.",
		Kind:    KindInt,
		Default: 2,
		Min:     ptr(0),
		Apply:   ApplyLive,
	})
	SwitchSNMPUser = register[string](Option{
		Key:     "switch.snmp.user",
		Section: "switch",
		Group:   "SNMP",
		Summary: "SNMPv3 user",
		Help:    "Username for SNMP version 3.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
	})
	SwitchSNMPAuthProto = register[string](Option{
		Key:     "switch.snmp.auth_proto",
		Section: "switch",
		Group:   "SNMP",
		Summary: "SNMPv3 authentication protocol",
		Help:    "How SNMP version 3 requests are authenticated. Empty sends them unauthenticated.",
		Kind:    KindEnum,
		Enum:    []string{"MD5", "SHA", "SHA224", "SHA256", "SHA384", "SHA512"},
		Default: "",
		Apply:   ApplyLive,
	})
	SwitchSNMPAuthPass = register[string](Option{
		Key:     "switch.snmp.auth_pass",
		Section: "switch",
		Group:   "SNMP",
		Summary: "SNMPv3 authentication password",
		Help:    "Password for the authentication protocol.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
		Secret:  true,
	})
	SwitchSNMPPrivProto = register[string](Option{
		Key:     "switch.snmp.priv_proto",
		Section: "switch",
		Group:   "SNMP",
		Summary: "SNMPv3 privacy protocol",
		Help:    "How SNMP version 3 requests are encrypted. Empty sends them unencrypted.",
		Kind:    KindEnum,
		Enum:    []string{"DES", "AES", "AES192", "AES256", "AES192C", "AES256C"},
		Default: "",
		Apply:   ApplyLive,
	})
	SwitchSNMPPrivPass = register[string](Option{
		Key:     "switch.snmp.priv_pass",
		Section: "switch",
		Group:   "SNMP",
		Summary: "SNMPv3 privacy password",
		Help:    "Password for the privacy protocol.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
		Secret:  true,
	})
)
