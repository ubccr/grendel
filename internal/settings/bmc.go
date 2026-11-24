// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import "time"

var (
	BMCUsername = register[string](Option{
		Key:      "bmc.username",
		Section:  "bmc",
		Summary:  "Username",
		Help:     "A single BMC login, tried before the ones in bmc.credentials. It suits the environment, as GRENDEL_BMC_USERNAME and GRENDEL_BMC_PASSWORD.",
		Kind:     KindString,
		Default:  "",
		Apply:    ApplyLive,
		External: true,
	})
	BMCUser = register[string](Option{
		Key:        "bmc.user",
		Section:    "bmc",
		Summary:    "Username",
		Help:       "Read when bmc.username is not set.",
		Kind:       KindString,
		Default:    "",
		Apply:      ApplyLive,
		External:   true,
		Deprecated: "use bmc.username",
		ReplacedBy: "bmc.username",
	})
	BMCPassword = register[string](Option{
		Key:     "bmc.password",
		Section: "bmc",
		Summary: "Password",
		Help:    "Password for bmc.username.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
		Secret:  true,
	})
	BMCCredentials = register[[]map[string]any](Option{
		Key:     "bmc.credentials",
		Section: "bmc",
		Summary: "Logins",
		Help:    "BMC logins, tried in order until one is accepted. List the current login first, then old passwords or factory defaults.",
		Kind:    KindTable,
		Default: []map[string]any{},
		Apply:   ApplyLive,
		Secret:  true,
		Fields: []Field{
			{Name: "username", Summary: "Username", Kind: KindString},
			{Name: "password", Summary: "Password", Kind: KindString},
		},
	})
	BMCDefaultCredentials = register[[]map[string]any](Option{
		Key:     "bmc.defaults.credentials",
		Section: "bmc",
		Summary: "Fallback logins",
		Help:    "Logins tried after the ones in bmc.credentials.",
		Kind:    KindTable,
		Default: []map[string]any{},
		Apply:   ApplyLive,
		Secret:  true,
		Fields: []Field{
			{Name: "user", Summary: "Username", Kind: KindString},
			{Name: "password", Summary: "Password", Kind: KindString},
		},
		Deprecated: "list them under bmc.credentials",
	})
	BMCInsecure = register[bool](Option{
		Key:     "bmc.insecure",
		Section: "bmc",
		Summary: "Accept untrusted certificates",
		Help:    "Talk to BMCs whose HTTPS certificate is self-signed or does not match. Most BMCs ship that way.",
		Kind:    KindBool,
		Default: false,
		Apply:   ApplyLive,
	})
	BMCScheme = register[string](Option{
		Key:     "bmc.scheme",
		Section: "bmc",
		Summary: "Scheme",
		Help:    "How BMCs are reached. http sends BMC passwords in plaintext and is meant only for a development Redfish mockup server.",
		Kind:    KindEnum,
		Enum:    []string{"https", "http"},
		Default: "https",
		Apply:   ApplyLive,
	})

	BMCFanout = register[int](Option{
		Key:     "bmc.fanout",
		Section: "bmc",
		Group:   "Running a command over many BMCs",
		Summary: "Parallel BMCs",
		Help:    "How many BMCs are talked to at once.",
		Kind:    KindInt,
		Default: 5,
		Min:     ptr(1),
		Apply:   ApplyLive,
	})
	BMCNodeTimeout = register[time.Duration](Option{
		Key:     "bmc.node_timeout",
		Section: "bmc",
		Group:   "Running a command over many BMCs",
		Summary: "Timeout per BMC",
		Help:    "How long one BMC may take before it is counted as failed.",
		Kind:    KindDuration,
		Default: 60 * time.Second,
		Apply:   ApplyLive,
	})
	BMCRunTimeout = register[time.Duration](Option{
		Key:     "bmc.run_timeout",
		Section: "bmc",
		Group:   "Running a command over many BMCs",
		Summary: "Timeout per run",
		Help:    "How long a whole run over a set of nodes may take. 0 sets no limit.",
		Kind:    KindDuration,
		Default: 4 * time.Minute,
		Apply:   ApplyLive,
	})
	BMCRateLimitPerSecond = register[float64](Option{
		Key:     "bmc.rate_limit_per_second",
		Section: "bmc",
		Group:   "Running a command over many BMCs",
		Summary: "New BMCs per second",
		Help:    "How many BMCs a run starts contacting each second. 0 sets no limit.",
		Kind:    KindFloat,
		Default: float64(0),
		Apply:   ApplyLive,
	})

	BMCFirmwarePollInterval = register[time.Duration](Option{
		Key:     "bmc.firmware_poll_interval",
		Section: "bmc",
		Group:   "Firmware updates",
		Summary: "Poll interval",
		Help:    "After a firmware update is submitted, how often the BMC job queues are read until it finishes.",
		Kind:    KindDuration,
		Default: time.Minute,
		Apply:   ApplyRestart,
	})
	BMCFirmwareTimeout = register[time.Duration](Option{
		Key:     "bmc.firmware_timeout",
		Section: "bmc",
		Group:   "Firmware updates",
		Summary: "Stalled after",
		Help:    "A node whose BMC has not reported on its firmware update for this long is marked stalled.",
		Kind:    KindDuration,
		Default: 2 * time.Hour,
		Apply:   ApplyLive,
	})

	BMCConfigShareIP = register[string](Option{
		Key:     "bmc.config_share_ip",
		Section: "bmc",
		Group:   "Dell system config import",
		Summary: "Share address",
		Help:    "IP address or hostname the iDRAC fetches an imported config from. It should be one of the provision server's that BMCs can reach. Empty uses the address the provision server listens on.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
	})
	BMCConfigIgnoreCertificateWarning = register[string](Option{
		Key:     "bmc.config_ignore_certificate_warning",
		Section: "bmc",
		Group:   "Dell system config import",
		Summary: "Ignore certificate warning",
		Help:    "Set to Enabled when the provision server uses a self-signed certificate.",
		Kind:    KindEnum,
		Enum:    []string{"Disabled", "Enabled"},
		Default: "Disabled",
		Apply:   ApplyLive,
	})

	BMCAssetTag = register[string](Option{
		Key:     "bmc.asset_tag",
		Section: "bmc",
		Group:   "Identity templates",
		Summary: "Asset tag",
		Help:    "Template for a node's asset tag when none is given, such as UB-{{ .Name }}. Templates see .Name, .Parts (the name split on -) and .Tags (key=value tags by key).",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
	})
	BMCLocationDatacenter = location("datacenter", "Datacenter")
	BMCLocationRoom       = location("room", "Room")
	BMCLocationRow        = location("row", "Row")
	BMCLocationRack       = location("rack", "Rack")
	BMCLocationRackUnit   = location("rack_unit", "Rack unit")

	BMCIPMIPort = register[int](Option{
		Key:     "bmc.ipmi.port",
		Section: "bmc",
		Group:   "Serial console (IPMI)",
		Summary: "Port",
		Help:    "UDP port for IPMI over LAN. The console needs IPMI over LAN and SOL enabled on the BMC, and this port reachable from the server.",
		Kind:    KindInt,
		Default: 623,
		Min:     ptr(1),
		Max:     ptr(65535),
		Apply:   ApplyLive,
	})
	BMCIPMICipherSuites = register[[]int](Option{
		Key:     "bmc.ipmi.cipher_suites",
		Section: "bmc",
		Group:   "Serial console (IPMI)",
		Summary: "Cipher suites",
		Help:    "RMCP+ cipher suites to try, in order. 17 is the iDRAC9 default and 3 covers older BMCs.",
		Kind:    KindIntList,
		Default: []int{17, 3},
		Apply:   ApplyLive,
	})

	BMCGofishBasicAuth = register[bool](Option{
		Key:     "bmc.gofish.basic_auth",
		Section: "bmc",
		Group:   "Redfish connection",
		Summary: "Basic auth",
		Help:    "Send the login with every request in place of opening a Redfish session. A mockup server needs this.",
		Kind:    KindBool,
		Default: true,
		Apply:   ApplyLive,
	})
	BMCGofishMaxConcurrentRequests = register[int](Option{
		Key:     "bmc.gofish.max_concurrent_requests",
		Section: "bmc",
		Group:   "Redfish connection",
		Summary: "Requests at once per BMC",
		Help:    "How many requests may be in flight to one BMC. Many BMCs misbehave above 1.",
		Kind:    KindInt,
		Default: 1,
		Min:     ptr(1),
		Apply:   ApplyLive,
	})
	BMCGofishReuseConnections = register[bool](Option{
		Key:     "bmc.gofish.reuse_connections",
		Section: "bmc",
		Group:   "Redfish connection",
		Summary: "Reuse connections",
		Help:    "Keep the connection to a BMC open between requests.",
		Kind:    KindBool,
		Default: true,
		Apply:   ApplyLive,
	})
	BMCGofishAutoExpand = register[bool](Option{
		Key:     "bmc.gofish.auto_expand",
		Section: "bmc",
		Group:   "Redfish connection",
		Summary: "Expand collections",
		Help:    "Read a collection and its members in one request on BMCs that support it, in place of one request per member. A BMC that fails the expanded read is asked again without it.",
		Kind:    KindBool,
		Default: true,
		Apply:   ApplyLive,
	})
	BMCGofishClientTimeout        = gofishTimeout("client_timeout", "Request timeout", "How long one request to a BMC may take.", 60*time.Second)
	BMCGofishDialTimeout          = gofishTimeout("dial_timeout", "Connect timeout", "How long opening a connection to a BMC may take.", 5*time.Second)
	BMCGofishDialKeepAliveTimeout = gofishTimeout("dial_keep_alive_timeout", "Keep-alive interval", "How often an idle connection is probed to check the BMC is still there.", 30*time.Second)
	BMCGofishTLSHandshakeTimeout  = gofishTimeout("tls_handshake_timeout", "TLS handshake timeout", "How long the TLS handshake with a BMC may take.", 10*time.Second)
	BMCGofishIdleConnTimeout      = gofishTimeout("idle_conn_timeout", "Idle connection timeout", "How long an unused connection to a BMC is kept open.", 90*time.Second)

	BMCSwitchAdminUsername = register[string](Option{
		Key:        "bmc.switch_admin_username",
		Section:    "switch",
		Summary:    "Old switch username",
		Help:       "Used when switch.username is empty.",
		Kind:       KindString,
		Default:    "",
		Apply:      ApplyLive,
		Deprecated: "use switch.username",
		ReplacedBy: "switch.username",
	})
	BMCSwitchAdminPassword = register[string](Option{
		Key:        "bmc.switch_admin_password",
		Section:    "switch",
		Summary:    "Old switch password",
		Help:       "Used when switch.password is empty.",
		Kind:       KindString,
		Default:    "",
		Apply:      ApplyLive,
		Secret:     true,
		Deprecated: "use switch.password",
		ReplacedBy: "switch.password",
	})
)

func location(name, summary string) Opt[string] {
	return register[string](Option{
		Key:     "bmc.location." + name,
		Section: "bmc",
		Group:   "Identity templates",
		Summary: summary,
		Help:    "Template for this part of a node's location, rendered per node. With names like cpn-f01-04, {{ index .Parts 1 }} gives the rack. Empty leaves it unwritten.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
	})
}

func gofishTimeout(name, summary, help string, def time.Duration) Opt[time.Duration] {
	return register[time.Duration](Option{
		Key:     "bmc.gofish." + name,
		Section: "bmc",
		Group:   "Redfish connection",
		Summary: summary,
		Help:    help,
		Kind:    KindDuration,
		Default: def,
		Apply:   ApplyLive,
	})
}
