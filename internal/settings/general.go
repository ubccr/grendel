// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"strings"
	"time"
)

var (
	DBType = register[string](Option{
		Key:      "dbtype",
		Section:  "general",
		Summary:  "Database type",
		Help:     "The database backend. It is needed before the database can be opened, so it cannot be stored there.",
		Kind:     KindEnum,
		Enum:     []string{"sqlite"},
		Default:  "sqlite",
		External: true,
		Flag:     "dbtype",
	})
	DBPath = register[string](Option{
		Key:      "dbpath",
		Packaged: "/var/lib/grendel/db/grendel.sqlite3",
		Section:  "general",
		Summary:  "Database path",
		Help:     "Where the database file lives. The default keeps everything in memory and loses it when grendeld stops. It is needed before the database can be opened, so it cannot be stored there.",
		Kind:     KindPath,
		Default:  ":memory:",
		External: true,
		Flag:     "dbpath",
	})
	DatabaseMaxReadConns = register[int](Option{
		Key:      "database.max_read_conns",
		Section:  "general",
		Summary:  "Database read connections",
		Help:     "Most reads that may run at once. Reads do not block each other, so there is little to gain above the core count, and an unbounded pool lets a burst of slow queries exhaust file descriptors.",
		Kind:     KindInt,
		Default:  8,
		Min:      ptr(1),
		External: true,
	})
	Services = register[[]string](Option{
		Key:      "services",
		Section:  "general",
		Summary:  "Services to run",
		Help:     "Which of api, dhcp, dns, provision, pxe and tftp this process runs. Empty runs them all. Naming services on the command line, as in `grendeld serve dhcp dns`, wins over this. It differs per process, so it is not stored in the database.",
		Kind:     KindStringList,
		Default:  []string{},
		External: true,
		Flag:     "services",
	})
	ShutdownTimeout = register[time.Duration](Option{
		Key:     "shutdown_timeout",
		Section: "general",
		Summary: "Shutdown timeout",
		Help:    "How long to wait for a service to stop before giving up on it during shutdown.",
		Kind:    KindDuration,
		Default: 10 * time.Second,
		Apply:   ApplyLive,
		Flag:    "shutdown-timeout",
	})
)

var LogLevel = register[string](Option{
	Key:     "log_level",
	Section: "general",
	Group:   "Loggers",
	Summary: "Log level",
	Help:    "How much grendeld logs: warn for warnings and errors only, info to add routine messages, debug for everything. The --verbose and --debug flags set info and debug. A logger with a level of its own ignores this.",
	Kind:    KindEnum,
	Enum:    []string{"warn", "info", "debug"},
	Default: "warn",
	Apply:   ApplyLive,
})

var loggers = registerLoggers()

func registerLoggers() map[string]Opt[string] {
	out := map[string]Opt[string]{}
	for _, name := range []string{"api", "bmc", "cli", "db", "dhcp", "dns", "events", "migrate", "node", "provision", "pxe", "redfish", "store", "switch", "template", "tftp"} {
		out[name] = register[string](Option{
			Key:     "loggers." + name,
			Section: "general",
			Group:   "Loggers",
			Summary: name,
			Help:    "What the " + name + " logger writes: on follows log_level, off writes nothing, and warn, info or debug set a level for this logger alone.",
			Kind:    KindEnum,
			Enum:    []string{"on", "off", "warn", "info", "debug"},
			Default: "on",
			Apply:   ApplyLive,
		})
	}
	return out
}

// LoggerLevel is the level a logger writes at: off, warn, info or debug. A logger with no setting of its own follows log_level.
func LoggerLevel(name string) string {
	if o, ok := loggers[strings.ToLower(name)]; ok {
		if level := o.Get(); level != "on" {
			return level
		}
	}
	return LogLevel.Get()
}

// MostVerboseLogLevel is the most detailed level any logger writes at, which is what the shared logger has to let through
func MostVerboseLogLevel() string {
	rank := map[string]int{"warn": 1, "info": 2, "debug": 3}
	most := LogLevel.Get()
	for name := range loggers {
		if level := LoggerLevel(name); rank[level] > rank[most] {
			most = level
		}
	}
	return most
}
