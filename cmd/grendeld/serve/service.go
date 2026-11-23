// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

type Runner interface {
	Serve() error
	Shutdown(context.Context) error
}

type Service struct {
	Name string
	New  func() (Runner, error)
}

var registry []*Service

func register(s *Service) {
	registry = append(registry, s)
}

func services() []*Service {
	return registry
}

func serviceNames(from []*Service) []string {
	names := make([]string, 0, len(from))
	for _, s := range from {
		names = append(names, s.Name)
	}

	return names
}

// selectServices resolves service names against the registry. An empty list means every service
func selectServices(names []string) ([]*Service, error) {
	selected := make([]*Service, 0, len(names))

	for _, arg := range names {
		// --services is a string slice and splits on commas, so accept the same form as a positional argument rather than making one comma an unknown service
		for _, name := range strings.Split(arg, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}

			i := slices.IndexFunc(registry, func(s *Service) bool { return s.Name == name })
			if i < 0 {
				return nil, fmt.Errorf("unknown service %q, valid services are: %s", name, strings.Join(serviceNames(registry), ", "))
			}

			if !slices.Contains(selected, registry[i]) {
				selected = append(selected, registry[i])
			}
		}
	}

	if len(selected) == 0 {
		return slices.Clone(registry), nil
	}

	return selected, nil
}
