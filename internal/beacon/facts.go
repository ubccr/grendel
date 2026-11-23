// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package beacon

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

// FactSource is where Collect looks for facts, so tests can point it somewhere else
type FactSource struct {
	OSRelease []string
	Kernel    string
}

// DefaultFacts reads the running system
var DefaultFacts = FactSource{
	OSRelease: []string{"/etc/os-release", "/usr/lib/os-release"},
	Kernel:    "/proc/sys/kernel/osrelease",
}

// Collect gathers the built in facts: os.* from the first os-release file there is, and kernel.release
func Collect(src FactSource) map[string]string {
	facts := map[string]string{}

	for _, path := range src.OSRelease {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for k, v := range parseOSRelease(data) {
			facts["os."+strings.ToLower(k)] = v
		}
		break
	}

	if data, err := os.ReadFile(src.Kernel); err == nil {
		facts["kernel.release"] = strings.TrimSpace(string(data))
	}

	return facts
}

// parseOSRelease reads os-release's KEY=value lines, where a value may be quoted and escaped the way a shell would
func parseOSRelease(data []byte) map[string]string {
	out := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || k == "" {
			continue
		}
		out[k] = unquote(v)
	}
	return out
}

func unquote(v string) string {
	if len(v) >= 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
		return v[1 : len(v)-1]
	}
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		v = v[1 : len(v)-1]
		var b strings.Builder
		for i := 0; i < len(v); i++ {
			if v[i] == '\\' && i+1 < len(v) && strings.IndexByte("\\$\"`", v[i+1]) >= 0 {
				i++
			}
			b.WriteByte(v[i])
		}
		return b.String()
	}
	return v
}
