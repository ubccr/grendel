// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package beacon

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	// DefaultFactsDir holds site facts: executables printing key=value lines, or plain files of them
	DefaultFactsDir = "/etc/grendel/beacon.d"

	factScriptTimeout = 10 * time.Second
	maxFactOutput     = 64 << 10
)

var (
	// a drop-in's name becomes its facts' prefix, so it has to be a valid key part. This also skips editor backups and package manager leftovers like slurm.rpmsave
	dropInName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
	factKey    = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,63}$`)
)

// FactSource is where Collect looks for facts, so tests can point it somewhere else
type FactSource struct {
	OSRelease []string
	Kernel    string
	DropIns   string
}

// DefaultFacts reads the running system
var DefaultFacts = FactSource{
	OSRelease: []string{"/etc/os-release", "/usr/lib/os-release"},
	Kernel:    "/proc/sys/kernel/osrelease",
	DropIns:   DefaultFactsDir,
}

// Collect gathers the built in facts, os.* from os-release, kernel.release and, once the boot has finished, boot.* timings, then each drop-in's facts under its own name. A drop-in can't replace a built in fact
func Collect(ctx context.Context, src FactSource, sys System) map[string]string {
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

	if sys != nil {
		if times, err := sys.BootTimes(ctx); err == nil {
			for k, v := range bootFacts(times) {
				facts[k] = v
			}
		}
	}

	for name, dropIn := range dropInFacts(ctx, src.DropIns) {
		for k, v := range dropIn {
			key := name + "." + k
			if _, ok := facts[key]; ok {
				logf("fact %s from %s ignored, it is already set", key, name)
				continue
			}
			facts[key] = v
		}
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

// bootFacts splits the boot the way systemd-analyze does, from the manager's monotonic timestamps in microseconds. Nothing is reported until the boot has finished
func bootFacts(t map[string]uint64) map[string]string {
	finish := t["FinishTimestampMonotonic"]
	if finish == 0 {
		return nil
	}
	firmware, loader := t["FirmwareTimestampMonotonic"], t["LoaderTimestampMonotonic"]
	initrd, userspace := t["InitRDTimestampMonotonic"], t["UserspaceTimestampMonotonic"]

	us := func(v uint64) string {
		return (time.Duration(v) * time.Microsecond).Round(time.Millisecond).String()
	}

	out := map[string]string{}
	if firmware > loader {
		out["boot.firmware"] = us(firmware - loader)
	}
	if loader > 0 {
		out["boot.loader"] = us(loader)
	}
	kernelEnd := userspace
	if initrd > 0 {
		kernelEnd = initrd
		if userspace > initrd {
			out["boot.initrd"] = us(userspace - initrd)
		}
	}
	if kernelEnd > 0 {
		out["boot.kernel"] = us(kernelEnd)
	}
	if finish > userspace {
		out["boot.userspace"] = us(finish - userspace)
	}
	out["boot.total"] = us(firmware + finish)
	return out
}

// dropInFacts reads each drop-in in dir: an executable is run and its output read, any other file is read as it is. Both hold key=value lines
func dropInFacts(ctx context.Context, dir string) map[string]map[string]string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logf("failed to read %s: %s", dir, err)
		}
		return nil
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	out := map[string]map[string]string{}
	for _, name := range names {
		if !dropInName.MatchString(name) {
			continue
		}
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}

		var data []byte
		if info.Mode()&0o111 != 0 {
			data, err = runDropIn(ctx, path)
		} else {
			data, err = os.ReadFile(path)
		}
		if err != nil {
			logf("facts from %s failed: %s", path, err)
			continue
		}

		facts := map[string]string{}
		sc := bufio.NewScanner(bytes.NewReader(data))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			k = strings.TrimSpace(k)
			if !ok || !factKey.MatchString(name+"."+k) {
				logf("skipping %q from %s, facts are key=value with a lowercase key", line, path)
				continue
			}
			facts[k] = strings.TrimSpace(v)
		}
		out[name] = facts
	}
	return out
}

func runDropIn(ctx context.Context, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, factScriptTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(stdout, maxFactOutput))
	// whatever is past the limit is dropped, and reading it lets the script finish rather than block on a full pipe
	_, _ = io.Copy(io.Discard, stdout)
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return data, readErr
}
