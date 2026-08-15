// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"slices"
	"strings"
	"testing"
)

// the fixtures below name these directly, so say so plainly rather than failing on a confusing mismatch if one is ever renamed
func requireRegistered(t *testing.T, names ...string) {
	t.Helper()

	all := serviceNames(services())
	for _, name := range names {
		if !slices.Contains(all, name) {
			t.Fatalf("service %q is not registered, the registry holds %q", name, all)
		}
	}
}

func TestSelectServices(t *testing.T) {
	requireRegistered(t, "api", "dns", "tftp")
	all := serviceNames(services())

	tests := []struct {
		name  string
		names []string
		want  []string
	}{
		{"no names is every service", nil, all},
		{"empty slice is every service", []string{}, all},
		{"blank entries are every service", []string{"", "  "}, all},
		{"one name", []string{"dns"}, []string{"dns"}},
		{"a subset keeps the order given", []string{"tftp", "api"}, []string{"tftp", "api"}},
		{"duplicates collapse", []string{"dns", "dns", "api", "dns"}, []string{"dns", "api"}},
		{"surrounding space is trimmed", []string{" dns ", "api"}, []string{"dns", "api"}},
		{"one comma separated argument", []string{"dns,api"}, []string{"dns", "api"}},
		{"commas mixed with separate arguments", []string{"dns,api", "tftp"}, []string{"dns", "api", "tftp"}},
		{"duplicates collapse across forms", []string{"dns,api", "dns"}, []string{"dns", "api"}},
		{"trailing comma", []string{"dns,"}, []string{"dns"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectServices(tt.names)
			if err != nil {
				t.Fatalf("selectServices(%q) returned an error: %s", tt.names, err)
			}

			if !slices.Equal(serviceNames(got), tt.want) {
				t.Errorf("selectServices(%q) = %q, want %q", tt.names, serviceNames(got), tt.want)
			}
		})
	}
}

func TestSelectServicesUnknownName(t *testing.T) {
	requireRegistered(t, "dns")

	_, err := selectServices([]string{"dns", "dnss"})
	if err == nil {
		t.Fatal("selectServices accepted an unknown service name")
	}

	// the message has to name the bad entry and the valid set, it is the only feedback the user gets
	if !strings.Contains(err.Error(), `"dnss"`) {
		t.Errorf("error does not name the unknown service: %s", err)
	}

	for _, name := range serviceNames(services()) {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error does not list the valid service %q: %s", name, err)
		}
	}
}

// a selection that fails partway must not leave the caller with the services it had resolved so far
func TestSelectServicesUnknownNameReturnsNothing(t *testing.T) {
	got, err := selectServices([]string{"dns", "dnss"})
	if err == nil {
		t.Fatal("selectServices accepted an unknown service name")
	}

	if got != nil {
		t.Errorf("selectServices returned %q alongside an error, want nil", serviceNames(got))
	}
}

// the runner selects by name, so a typo or a copied entry in the registry would silently start the wrong service
func TestRegistryNamesAreUnique(t *testing.T) {
	names := serviceNames(services())
	if len(names) == 0 {
		t.Fatal("no services registered")
	}

	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if seen[name] {
			t.Errorf("service %q is registered more than once", name)
		}
		seen[name] = true
	}
}

func TestRegistryIsComplete(t *testing.T) {
	for _, svc := range services() {
		if svc.New == nil {
			t.Errorf("service %q has no New", svc.Name)
		}
	}
}
