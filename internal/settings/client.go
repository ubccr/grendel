// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

var (
	ClientAPIEndpoint = register[string](Option{
		Key:     "client.api_endpoint",
		Summary: "API endpoint",
		Help:    "Where the grendel command finds the API: a URL such as http://localhost:8080, or the path of the server's unix socket.",
		Kind:    KindString,
		Default: "grendel-api.socket",
		Scope:   ScopeClient,
		Apply:   ApplyLive,
		Flag:    "endpoint",
	})
	ClientAPIKey = register[string](Option{
		Key:     "client.api_key",
		Summary: "API token",
		Help:    "Token the grendel command signs in with over TCP. Create one with `grendel auth token`.",
		Kind:    KindString,
		Default: "",
		Scope:   ScopeClient,
		Apply:   ApplyLive,
		Secret:  true,
	})
	ClientInsecure = register[bool](Option{
		Key:     "client.insecure",
		Summary: "Skip certificate checks",
		Help:    "Accept any certificate the API serves. Only for a server with a self-signed certificate that you trust.",
		Kind:    KindBool,
		Default: false,
		Scope:   ScopeClient,
		Apply:   ApplyLive,
	})
	ClientCACert = register[string](Option{
		Key:     "client.cacert",
		Summary: "CA certificate",
		Help:    "Path to a CA certificate to trust when checking the API's certificate.",
		Kind:    KindPath,
		Default: "",
		Scope:   ScopeClient,
		Apply:   ApplyLive,
	})
)
