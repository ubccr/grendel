// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

var (
	APIListen = register[string](Option{
		Key:      "api.listen",
		Section:  "api",
		Summary:  "Listen address",
		Help:     "Address and TCP port the API binds to, such as 0.0.0.0:8080. Requests over TCP are authenticated and checked against roles. Bind to an internal management network where possible. Empty serves the API on the unix socket only.",
		Kind:     KindString,
		Validate: validAddr,
		Default:  "",
		Apply:    ApplyRestart,
		Flag:     "api-listen",
	})
	APISocketPath = register[string](Option{
		Key:      "api.socket_path",
		Packaged: "grendel-api.socket",
		Section:  "api",
		Summary:  "Unix socket path",
		Help:     "Path of a unix socket the API also listens on. Requests over it skip authentication, so its file permissions decide who may use it. It is a path on the server, so it is not stored in the database.",
		Kind:     KindPath,
		Default:  "",
		External: true,
		Flag:     "api-socket",
	})
	APISecret = register[string](Option{
		Key:            "api.secret",
		Section:        "api",
		Summary:        "Signing secret",
		Help:           "Signs login sessions and API tokens. When nothing sets it, one is generated and kept in the database. Changing it signs everyone out. Generate one with `openssl rand -hex 16`.",
		Kind:           KindString,
		Default:        "",
		Apply:          ApplyLive,
		Secret:         true,
		Generated:      true,
		GeneratedBytes: 32,
	})
	APICert = register[string](Option{
		Key:      "api.cert",
		Section:  "api",
		Summary:  "TLS certificate",
		Help:     "Path to the certificate (.crt) the API serves. Setting both the certificate and the key turns on HTTPS, which you should always use over TCP.",
		Kind:     KindPath,
		Default:  "",
		External: true,
		Flag:     "api-cert",
	})
	APIKey = register[string](Option{
		Key:      "api.key",
		Section:  "api",
		Summary:  "TLS key",
		Help:     "Path to the private key (.key) for the API certificate.",
		Kind:     KindPath,
		Default:  "",
		External: true,
		Flag:     "api-key",
	})
	APIWebUI = register[bool](Option{
		Key:     "api.web_ui",
		Section: "api",
		Summary: "Web UI",
		Help:    "Serve this web UI at /ui. It is only served on the TCP listener.",
		Kind:    KindBool,
		Default: false,
		Apply:   ApplyRestart,
	})
	APISwaggerUI = register[bool](Option{
		Key:     "api.swagger_ui",
		Section: "api",
		Group:   "Development",
		Summary: "Swagger UI",
		Help:    "Serve an API browser at /swagger. Trying routes from it also needs CORS turned on.",
		Kind:    KindBool,
		Default: false,
		Apply:   ApplyRestart,
	})
	APICORS = register[bool](Option{
		Key:     "api.cors",
		Section: "api",
		Group:   "Development",
		Summary: "CORS",
		Help:    "Allow any request method from any origin. Leave this off outside development.",
		Kind:    KindBool,
		Default: false,
		Apply:   ApplyRestart,
	})
	APIDev = register[bool](Option{
		Key:      "api.dev",
		Section:  "api",
		Group:    "Development",
		Summary:  "Skip role checks",
		Help:     "Lets every signed-in account call every route. It turns access control off, so it can only be set in the config file or the environment.",
		Kind:     KindBool,
		Default:  false,
		Apply:    ApplyLive,
		External: true,
	})
)
