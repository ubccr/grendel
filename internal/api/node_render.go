// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"slices"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/config"
	"github.com/ubccr/grendel/internal/render"
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/model"
)

type NodeRenderResult struct {
	Name     string `json:"name"`
	Arch     string `json:"arch" description:"Arch the node was rendered as: its own, or the one asked for"`
	Image    string `json:"image" description:"Boot image the node was rendered with: its own, or the default"`
	Template string `json:"template" description:"Template the output came from" example:"ipxe.tmpl"`
	Content  string `json:"content" description:"What the provision server would send the node. Empty when the render failed"`
	Error    string `json:"error" description:"Why the node could not be rendered. Empty when it rendered"`
}

type NodeRenderResponse struct {
	Nodes []NodeRenderResult `json:"nodes"`
}

type NodeCheckFile struct {
	Name  string `json:"name" description:"File as the provision server serves it" example:"initrd-0"`
	Path  string `json:"path" description:"Path on the server, after the image's path template is rendered for the node"`
	Size  int64  `json:"size" description:"Size in bytes"`
	Error string `json:"error" description:"Why the path could not be rendered or the file read. Empty when it is readable"`
}

type NodeCheckTemplate struct {
	Kind     string `json:"kind" description:"ipxe, or a key of the image's provision_templates" example:"kickstart"`
	Template string `json:"template" example:"kickstart.tmpl"`
	Error    string `json:"error" description:"Why the template did not render. Empty when it rendered"`
}

type NodeCheckOnRequest struct {
	Kind     string `json:"kind" description:"What to ask the render route for" example:"kickstart"`
	Template string `json:"template" example:"kickstart.tmpl"`
}

type NodeCheckResult struct {
	Name      string               `json:"name"`
	Arch      string               `json:"arch" description:"Arch the node was checked as: its own, or the one asked for"`
	Image     string               `json:"image" description:"Boot image the node was checked with: its own, or the default"`
	Ok        bool                 `json:"ok" description:"Every file is readable and every template rendered"`
	Error     string               `json:"error" description:"Why nothing could be checked, such as a node with no boot image"`
	Files     []NodeCheckFile      `json:"files"`
	Templates []NodeCheckTemplate  `json:"templates"`
	OnRequest []NodeCheckOnRequest `json:"on_request" description:"What the node is also served if it asks, from a default template the image does not replace. These are not checked"`
}

type NodeCheckResponse struct {
	Nodes []NodeCheckResult `json:"nodes"`
}

// renderTarget is one node ready to render as one arch, or the reason it can't be
type renderTarget struct {
	host  *model.Host
	image *model.BootImage
	data  map[string]any
	err   error
}

func (t renderTarget) imageName() string {
	if t.image == nil {
		return ""
	}
	return t.image.Name
}

// renderer builds template data for nodes as the provision server would for a request from each
type renderer struct {
	db  store.Store
	set *render.Set
	// server is the address rendered URLs point at
	server string
	bmc    bool
	// arches are what a node with no arch of its own is rendered as, one target each
	arches []string
	images map[string]*model.BootImage
}

func (h *Handler) newRenderer(c fuego.ContextNoBody) (*renderer, error) {
	r := &renderer{db: h.DB, arches: []string{c.QueryParam("arch")}, images: map[string]*model.BootImage{}}
	if r.arches[0] == "all" {
		r.arches = model.Arches
	}

	switch iface := c.QueryParam("interface"); iface {
	case "", "boot":
	case "bmc":
		r.bmc = true
	default:
		return nil, badRequest(fmt.Errorf("unknown interface %q", iface), "interface must be boot or bmc")
	}

	// read fresh rather than shared with the provision server, so a template that no longer parses is reported instead of hidden behind the last set that did
	var err error
	if r.set, err = render.ReadSet(); err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusUnprocessableEntity,
			Err:    err,
			Title:  "Error",
			Detail: fmt.Sprintf("the templates do not parse: %s", err),
		}
	}

	if r.server = c.QueryParam("server"); r.server == "" {
		listen := config.ProvisionAddr().Addr().String()
		if r.server, err = util.ServerAddress(listen); err != nil {
			return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to find the provision server's address, give one as server"}
		}
	}

	return r, nil
}

// image loads a node's boot image, or the default one for a node with none, once however many nodes share it
func (r *renderer) image(name string) (*model.BootImage, error) {
	if name == "" {
		name = settings.ProvisionDefaultImage.Get()
	}
	if name == "" {
		return nil, errors.New("no boot image is set on the node, and provision.default_image is not set")
	}
	if image, ok := r.images[name]; ok {
		return image, nil
	}

	image, err := r.db.LoadBootImage(name)
	if err != nil {
		return nil, fmt.Errorf("boot image %s: %w", name, err)
	}
	r.images[name] = image
	return image, nil
}

// prepare gathers what rendering host needs beyond the host itself: the interface it boots from, its image, and a boot token
func (r *renderer) prepare(host *model.Host) (*model.NetInterface, *model.BootImage, string, error) {
	nic := host.BootInterface()
	if r.bmc {
		nic = host.InterfaceBMC()
	}
	if nic == nil || len(nic.MAC) == 0 {
		return nil, nil, "", errors.New("the node has no such interface with a MAC address")
	}

	image, err := r.image(host.BootImage)
	if err != nil {
		return nil, nil, "", err
	}

	token, err := model.NewBootTokenWithArch(host.UID.String(), nic.MAC.String(), host.Arch)
	if err != nil {
		return nil, image, "", fmt.Errorf("failed to create a boot token: %w", err)
	}
	return nic, image, token, nil
}

// targets is every host as each arch it is rendered as. A manually set arch takes precedence, as it does over the one detected during a PXE boot
func (r *renderer) targets(hosts model.HostList) []renderTarget {
	targets := make([]renderTarget, 0, len(hosts))
	for _, stored := range hosts {
		arches := r.arches
		if stored.Arch != "" {
			arches = []string{stored.Arch}
		}

		for _, arch := range arches {
			host := *stored
			host.Arch = arch

			nic, image, token, err := r.prepare(&host)
			if err != nil {
				// none of what prepare gathers depends on the arch, so the node is listed once, as stored
				targets = append(targets, renderTarget{host: stored, image: image, err: err})
				break
			}

			targets = append(targets, renderTarget{
				host:  &host,
				image: image,
				data: render.Data(render.Request{
					Host:    &host,
					NIC:     nic,
					Image:   image,
					Server:  r.server,
					Token:   token,
					Headers: http.Header{},
				}),
			})
		}
	}
	return targets
}

// selectedHosts are the nodes a request's nodeset and tags pick, or every node when it gives neither and every is allowed
func (h *Handler) selectedHosts(c fuego.ContextNoBody, every bool) (model.HostList, error) {
	ns, err := h.selectedNodeSet(c)
	if err != nil {
		return nil, err
	}

	var hosts model.HostList
	switch {
	case ns != nil:
		hosts, err = h.DB.FindHosts(ns)
	case every:
		hosts, err = h.DB.Hosts()
	default:
		return nil, badRequest(errors.New("no node selector given"), "specify nodeset or tags")
	}
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to find nodes"}
	}
	if len(hosts) == 0 {
		return nil, notFound("no nodes found")
	}
	return hosts, nil
}

func (h *Handler) NodeRender(c fuego.ContextNoBody) (*NodeRenderResponse, error) {
	kind := c.PathParam("template")

	// the output of every node would be a boot token for each, so the nodes have to be named
	hosts, err := h.selectedHosts(c, false)
	if err != nil {
		return nil, err
	}
	r, err := h.newRenderer(c)
	if err != nil {
		return nil, err
	}
	targets := r.targets(hosts)

	res := &NodeRenderResponse{Nodes: make([]NodeRenderResult, 0, len(targets))}
	for _, t := range targets {
		out := NodeRenderResult{Name: t.host.Name, Arch: t.host.Arch, Image: t.imageName()}
		if t.err != nil {
			out.Error = t.err.Error()
			res.Nodes = append(res.Nodes, out)
			continue
		}

		name, content, err := r.set.Render(kind, t.image, t.data)
		out.Template = name
		if err != nil {
			out.Error = err.Error()
		} else {
			out.Content = string(content)
		}
		res.Nodes = append(res.Nodes, out)
	}

	return res, nil
}

func (h *Handler) NodeCheck(c fuego.ContextNoBody) (*NodeCheckResponse, error) {
	hosts, err := h.selectedHosts(c, true)
	if err != nil {
		return nil, err
	}
	r, err := h.newRenderer(c)
	if err != nil {
		return nil, err
	}
	targets := r.targets(hosts)

	// most nodes share an image and so its files, which are read once
	read := map[string]NodeCheckFile{}
	check := func(name, path string) NodeCheckFile {
		if f, ok := read[path]; ok {
			f.Name = name
			return f
		}

		f := NodeCheckFile{Path: path}
		if size, err := readableFile(path); err != nil {
			f.Error = err.Error()
		} else {
			f.Size = size
		}
		read[path] = f
		f.Name = name
		return f
	}

	res := &NodeCheckResponse{Nodes: make([]NodeCheckResult, 0, len(targets))}
	for _, t := range targets {
		out := NodeCheckResult{Name: t.host.Name, Arch: t.host.Arch, Image: t.imageName(), Files: []NodeCheckFile{}, Templates: []NodeCheckTemplate{}, OnRequest: []NodeCheckOnRequest{}}
		if t.err != nil {
			out.Error = t.err.Error()
			res.Nodes = append(res.Nodes, out)
			continue
		}

		out.Ok = true
		for _, file := range render.Files(t.image, t.data) {
			if file.Err != nil {
				out.Ok = false
				out.Files = append(out.Files, NodeCheckFile{Name: file.Name, Error: file.Err.Error()})
				continue
			}

			names := []string{""}
			if t.image.Verify {
				names = append(names, ".sig")
			}
			for _, suffix := range names {
				f := check(file.Name+suffix, file.Path+suffix)
				out.Ok = out.Ok && f.Error == ""
				out.Files = append(out.Files, f)
			}
		}

		kinds := []string{render.KindIPXE}
		for _, kind := range slices.Sorted(maps.Keys(t.image.ProvisionTemplates)) {
			if kind != render.KindIPXE {
				kinds = append(kinds, kind)
			}
		}
		for _, kind := range kinds {
			as := kind
			// a butane template is only ever served translated, and the translation is where most of its mistakes surface
			if kind == render.KindButane {
				as = render.KindIgnition
			}

			name, _, err := r.set.Render(as, t.image, t.data)
			tmpl := NodeCheckTemplate{Kind: kind, Template: name}
			if err != nil {
				out.Ok = false
				tmpl.Error = err.Error()
			}
			out.Templates = append(out.Templates, tmpl)
		}

		for _, tmpl := range r.set.OnRequest(t.image) {
			out.OnRequest = append(out.OnRequest, NodeCheckOnRequest{Kind: tmpl.Kind, Template: tmpl.Name})
		}

		res.Nodes = append(res.Nodes, out)
	}

	return res, nil
}

// readableFile is the size of the file at path, once it is known to be one the server can open to send
func readableFile(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	// anything else is not a boot file, and opening a pipe would wait on its writer
	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), f.Close()
}
