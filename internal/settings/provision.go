// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

var (
	ProvisionListen = register[string](Option{
		Key:      "provision.listen",
		Section:  "provision",
		Summary:  "Listen address",
		Help:     "Address and TCP port the provision server binds to. Hosts fetch their boot scripts, kernels and install configs from it.",
		Kind:     KindString,
		Validate: validAddr,
		Default:  "0.0.0.0:80",
		Required: true,
		Apply:    ApplyRestart,
		Flag:     "provision-listen",
	})
	ProvisionHostname = register[string](Option{
		Key:     "provision.hostname",
		Section: "provision",
		Summary: "Hostname",
		Help:    "Name hosts use to reach the provision server, in place of its IP address. It should match the TLS certificate and resolve in DNS.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
	})
	ProvisionCert = register[string](Option{
		Key:      "provision.cert",
		Section:  "provision",
		Summary:  "TLS certificate",
		Help:     "Path to the certificate (.crt) the provision server serves. Setting both the certificate and the key turns on HTTPS.",
		Kind:     KindPath,
		Default:  "",
		External: true,
		Flag:     "provision-cert",
	})
	ProvisionKey = register[string](Option{
		Key:      "provision.key",
		Section:  "provision",
		Summary:  "TLS key",
		Help:     "Path to the private key (.key) for the provision certificate.",
		Kind:     KindPath,
		Default:  "",
		External: true,
		Flag:     "provision-key",
	})
	ProvisionTokenTTL = register[int](Option{
		Key:     "provision.token_ttl",
		Section: "provision",
		Summary: "Boot token lifetime",
		Help:    "How long a boot token handed to a host stays valid. A host that takes longer than this to install cannot fetch the rest of its files. 0 makes tokens that never expire.",
		Kind:    KindInt,
		Default: 3600,
		Unit:    "seconds",
		Min:     ptr(0),
		Apply:   ApplyLive,
	})
	ProvisionSecret = register[string](Option{
		Key:       "provision.secret",
		Section:   "provision",
		Summary:   "Boot token secret",
		Help:      "Signs the boot tokens handed to hosts, and must be exactly 32 characters, as `openssl rand -hex 16` gives. Left unset, one is generated and kept in the database.",
		Kind:      KindString,
		Default:   "",
		Apply:     ApplyLive,
		Secret:    true,
		Generated: true,
		// branca keys are exactly 32 characters
		GeneratedBytes: 16,
	})
	ProvisionDefaultImage = register[string](Option{
		Key:     "provision.default_image",
		Section: "provision",
		Summary: "Default image",
		Help:    "Name of the boot image used for a host that has none assigned.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyLive,
		Flag:    "default-image",
	})
	ProvisionRepoDir = register[string](Option{
		Key:      "provision.repo_dir",
		Section:  "provision",
		Summary:  "Repository directory",
		Help:     "Directory on this machine served to hosts under /repo during provisioning.",
		Kind:     KindPath,
		Default:  "",
		External: true,
		Flag:     "repo-dir",
	})
	ProvisionTemplatesDir = register[string](Option{
		Key:      "provision.templates_dir",
		Section:  "provision",
		Summary:  "Templates directory",
		Help:     "Directory of your own *.tmpl provision templates, which are loaded over the built-in ones and reloaded when they change.",
		Kind:     KindPath,
		Default:  "",
		External: true,
		Flag:     "templates-dir",
	})
	ProvisionStatusCheckSource = register[bool](Option{
		Key:     "provision.status_check_source",
		Section: "provision",
		Summary: "Check status report source",
		Help:    "Only accept a provision status report that comes from one of the host's own IP addresses.",
		Kind:    KindBool,
		Default: true,
		Apply:   ApplyLive,
	})
	ProvisionExtraVars = register[map[string]any](Option{
		Key:               "provision.extra_vars",
		Section:           "provision",
		Summary:           "Extra template variables",
		Help:              "Your own values passed to provision templates, read as {{ $.extra_vars.name }}. Entries in the config file win and are never shown, so keep passwords there: entries in the database can be read by anyone who can read settings.",
		Kind:              KindMap,
		Default:           map[string]any{},
		Apply:             ApplyLive,
		FileEntriesSecret: true,
	})

	ProvisionEnablePrometheusSD = register[bool](Option{
		Key:     "provision.enable_prometheus_sd",
		Section: "provision",
		Group:   "Integrations",
		Summary: "Prometheus service discovery",
		Help:    "Serve the hosts and PDUs grendel knows as Prometheus HTTP service discovery targets.",
		Kind:    KindBool,
		Default: false,
		Apply:   ApplyLive,
	})
	ProvisionPrometheusSDRefreshInterval = register[int](Option{
		Key:     "provision.prometheus_sd_refresh_interval",
		Section: "provision",
		Group:   "Integrations",
		Summary: "Service discovery refresh interval",
		Help:    "How often Prometheus is told to fetch the target list again.",
		Kind:    KindInt,
		Default: 3600,
		Unit:    "seconds",
		Min:     ptr(1),
		Apply:   ApplyLive,
	})
	ProvisionNetboxURL = register[string](Option{
		Key:     "provision.netbox_url",
		Section: "provision",
		Group:   "Integrations",
		Summary: "NetBox URL",
		Help:    "Base URL of a NetBox server to render switch configs from. NetBox rendering needs both this and the token.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyRestart,
	})
	ProvisionNetboxToken = register[string](Option{
		Key:     "provision.netbox_token",
		Section: "provision",
		Group:   "Integrations",
		Summary: "NetBox token",
		Help:    "API token grendel sends to NetBox.",
		Kind:    KindString,
		Default: "",
		Apply:   ApplyRestart,
		Secret:  true,
	})

	ProvisionBeaconEnabled = register[bool](Option{
		Key:     "provision.beacon.enabled",
		Section: "provision",
		Group:   "Beacon",
		Summary: "Install the beacon",
		Help:    "Add the beacon, which reports each boot back to grendel, to the default butane template.",
		Kind:    KindBool,
		Default: false,
		Apply:   ApplyLive,
	})
	ProvisionBeaconDir = register[string](Option{
		Key:      "provision.beacon.dir",
		Section:  "provision",
		Group:    "Beacon",
		Summary:  "Beacon binaries directory",
		Help:     "Directory holding grendel-beacon-amd64 and grendel-beacon-arm64, served at /beacon/<arch> for templates to download. The deb and rpm packages install both in /usr/lib/grendel/beacon.",
		Kind:     KindPath,
		Default:  "",
		External: true,
	})
	ProvisionBeaconPath = register[string](Option{
		Key:     "provision.beacon.path",
		Section: "provision",
		Group:   "Beacon",
		Summary: "Beacon path on the host",
		Help:    "Where the beacon is installed on a provisioned host.",
		Kind:    KindString,
		Default: "/opt/bin/grendel-beacon",
		Apply:   ApplyLive,
	})
	ProvisionBeaconComplete = register[bool](Option{
		Key:     "provision.beacon.complete",
		Section: "provision",
		Group:   "Beacon",
		Summary: "Finish provisioning on a clean boot",
		Help:    "When the beacon reports a clean boot, also turn provisioning off for the host, as the complete endpoint does.",
		Kind:    KindBool,
		Default: false,
		Apply:   ApplyLive,
	})
)
