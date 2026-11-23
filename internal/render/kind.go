// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package render

import (
	"bytes"
	"fmt"
	"maps"

	"github.com/ubccr/grendel/pkg/model"
)

const (
	KindIPXE      = "ipxe"
	KindKickstart = "kickstart"
	KindUserData  = "user_data"
	KindMetaData  = "meta_data"
	KindButane    = "butane"
	KindIgnition  = "ignition"
	KindProxmox   = "proxmox"
)

// MetaDataTemplate is served as cloud-init meta-data whatever the image
const MetaDataTemplate = "meta-data.tmpl"

var defaultTemplates = map[string]string{
	KindIPXE:      "ipxe.tmpl",
	KindKickstart: "kickstart.tmpl",
	KindUserData:  "user-data.tmpl",
	KindMetaData:  MetaDataTemplate,
	KindButane:    "butane.tmpl",
	KindProxmox:   "proxmox.tmpl",
}

// onRequest is what a node can ask for whatever its image, in the order a list shows them
var onRequest = []string{KindKickstart, KindUserData, KindMetaData, KindIgnition}

// TemplateFor is the template the image serves for kind: the one its provision_templates names, else the default for a kind that has one
func TemplateFor(image *model.BootImage, kind string) (string, bool) {
	if kind == KindMetaData {
		return MetaDataTemplate, true
	}
	if name, ok := image.ProvisionTemplates[kind]; ok {
		return name, true
	}

	name, ok := defaultTemplates[kind]
	return name, ok
}

// Render returns what a node is served for kind, and the template it came from. Kind is one of the Kind constants, a key of the image's provision_templates, or the name of a template in the set. data is the node's Data
func (s *Set) Render(kind string, image *model.BootImage, data map[string]any) (string, []byte, error) {
	if kind == KindIgnition {
		name, _ := TemplateFor(image, KindButane)
		out, err := s.Ignition(name, data)
		return name, out, err
	}

	name, ok := TemplateFor(image, kind)
	if !ok {
		if !s.Has(kind) {
			return "", nil, fmt.Errorf("%q is not a kind of template, a provision template of image %s, or a template file", kind, image.Name)
		}
		name = kind
	}

	if kind == KindIPXE {
		// the command line is the iPXE script's alone, so the caller's data is left as it was given
		data = maps.Clone(data)
		if err := AddCommandLine(image, data); err != nil {
			return name, nil, fmt.Errorf("kernel command line: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := s.Execute(&buf, name, data); err != nil {
		return name, nil, err
	}
	return name, buf.Bytes(), nil
}

// Template is a kind of output and the template it is rendered from
type Template struct {
	Kind string
	Name string
}

// OnRequest lists what a node booting image is served if it asks, beyond what the image's provision_templates names: the kinds with a default template that is in the set
func (s *Set) OnRequest(image *model.BootImage) []Template {
	var out []Template
	for _, kind := range onRequest {
		source := kind
		if kind == KindIgnition {
			source = KindButane
		}
		if _, named := image.ProvisionTemplates[source]; named {
			continue
		}
		if name := defaultTemplates[source]; s.Has(name) {
			out = append(out, Template{Kind: kind, Name: name})
		}
	}
	return out
}
