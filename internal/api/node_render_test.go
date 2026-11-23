// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
)

// renderFixture stores an image whose kernel path is templated on the node's arch, a node booting it, and a templates dir holding custom.tmpl
func renderFixture(t *testing.T, custom string) (*Handler, string) {
	t.Helper()
	h := newTestHandler(t)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "custom.tmpl"), []byte(custom), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "vmlinuz-x86_64"), []byte("kernel"), 0o644))
	viper.Set("provision.templates_dir", dir)
	t.Cleanup(func() { viper.Set("provision.templates_dir", nil) })

	image := &model.BootImage{
		Name:               "render-test",
		KernelPath:         filepath.Join(dir, "vmlinuz-{{ .host.Arch }}"),
		InitrdPaths:        []string{filepath.Join(dir, "initrd.img")},
		CommandLine:        "console=ttyS0 host={{ .host.Name }}",
		ProvisionTemplates: map[string]string{"post": "custom.tmpl"},
	}
	require.NoError(t, h.DB.StoreBootImage(image))

	host := &model.Host{
		Name:       "cpn-01",
		BootImage:  image.Name,
		Interfaces: []*model.NetInterface{{MAC: net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}}},
	}
	require.NoError(t, h.DB.StoreHost(host))

	return h, dir
}

func renderRequest(template string) *fuego.MockContext[any] {
	c := fuego.NewMockContextNoBody()
	c.PathParams["template"] = template
	c.SetQueryParam("nodeset", "cpn-01")
	c.SetQueryParam("server", "10.0.0.1")
	c.SetQueryParam("arch", "x86_64")
	return c
}

func TestNodeRender(t *testing.T) {
	h, _ := renderFixture(t, "hello {{ .host.Name }} {{ .endpoints.CompleteURL }}")

	res, err := h.NodeRender(renderRequest("ipxe"))
	require.NoError(t, err)
	require.Len(t, res.Nodes, 1)
	assert.Empty(t, res.Nodes[0].Error)
	assert.Equal(t, "ipxe.tmpl", res.Nodes[0].Template)
	assert.Equal(t, "render-test", res.Nodes[0].Image)
	assert.Contains(t, res.Nodes[0].Content, "#!ipxe")
	assert.Contains(t, res.Nodes[0].Content, "console=ttyS0 host=cpn-01")
	assert.Contains(t, res.Nodes[0].Content, "http://10.0.0.1/boot/")

	// a key of the image's provision_templates, and the file it names
	for _, name := range []string{"post", "custom.tmpl"} {
		res, err = h.NodeRender(renderRequest(name))
		require.NoError(t, err)
		assert.Empty(t, res.Nodes[0].Error)
		assert.Equal(t, "custom.tmpl", res.Nodes[0].Template)
		assert.Regexp(t, `^hello cpn-01 http://10\.0\.0\.1/boot/[^/]+/complete$`, res.Nodes[0].Content)
	}

	// cloud-init meta-data is the built-in template whatever the image names
	res, err = h.NodeRender(renderRequest("meta_data"))
	require.NoError(t, err)
	assert.Equal(t, "meta-data.tmpl", res.Nodes[0].Template)

	res, err = h.NodeRender(renderRequest("nope"))
	require.NoError(t, err)
	assert.Contains(t, res.Nodes[0].Error, `"nope" is not a kind of template`)

	host, err := h.DB.LoadHostFromName("cpn-01")
	require.NoError(t, err)
	_, err = h.DB.LoadProvisionStatus(host.UID.String())
	assert.ErrorIs(t, err, store.ErrNotFound, "rendering is not a stage of a provision")
}

func TestNodeRenderRefusals(t *testing.T) {
	h, _ := renderFixture(t, "ok")
	var he fuego.HTTPError

	every := fuego.NewMockContextNoBody()
	every.PathParams["template"] = "ipxe"
	_, err := h.NodeRender(every)
	require.ErrorAs(t, err, &he, "rendering every node would hand out a token for each")
	assert.Equal(t, http.StatusBadRequest, he.Status)

	c := renderRequest("ipxe")
	c.SetQueryParam("interface", "bcm")
	_, err = h.NodeRender(c)
	require.ErrorAs(t, err, &he)
	assert.Equal(t, "interface must be boot or bmc", he.Detail)
}

func TestNodeRenderReportsTemplateErrors(t *testing.T) {
	h, _ := renderFixture(t, "{{ .host.Nope }}")
	res, err := h.NodeRender(renderRequest("post"))
	require.NoError(t, err)
	assert.Contains(t, res.Nodes[0].Error, "Nope")
	assert.Empty(t, res.Nodes[0].Content)

	h, _ = renderFixture(t, "{{ if")
	_, err = h.NodeRender(renderRequest("ipxe"))
	var he fuego.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Contains(t, he.Detail, "custom.tmpl", "a template that does not parse fails the request, naming the file")
}

func TestNodeCheck(t *testing.T) {
	h, dir := renderFixture(t, "ok")

	res, err := h.NodeCheck(renderRequest(""))
	require.NoError(t, err)
	require.Len(t, res.Nodes, 1)
	node := res.Nodes[0]
	assert.False(t, node.Ok)

	require.Len(t, node.Files, 2)
	assert.Equal(t, NodeCheckFile{Name: "kernel", Path: filepath.Join(dir, "vmlinuz-x86_64"), Size: 6}, node.Files[0])
	assert.Equal(t, "initrd-0", node.Files[1].Name)
	assert.Contains(t, node.Files[1].Error, "no such file")

	assert.Equal(t, []NodeCheckTemplate{{Kind: "ipxe", Template: "ipxe.tmpl"}, {Kind: "post", Template: "custom.tmpl"}}, node.Templates)
	assert.Equal(t, []NodeCheckOnRequest{
		{Kind: "kickstart", Template: "kickstart.tmpl"},
		{Kind: "user_data", Template: "user-data.tmpl"},
		{Kind: "meta_data", Template: "meta-data.tmpl"},
		{Kind: "ignition", Template: "butane.tmpl"},
	}, node.OnRequest)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "initrd.img"), []byte("initrd"), 0o644))
	res, err = h.NodeCheck(renderRequest(""))
	require.NoError(t, err)
	assert.True(t, res.Nodes[0].Ok)

	// neither nodeset nor tags is every node
	every := fuego.NewMockContextNoBody()
	every.SetQueryParam("server", "10.0.0.1")
	every.SetQueryParam("arch", "x86_64")
	res, err = h.NodeCheck(every)
	require.NoError(t, err)
	require.Len(t, res.Nodes, 1)
	assert.True(t, res.Nodes[0].Ok)

	// a node with no arch of its own is checked as each arch
	both := renderRequest("")
	both.SetQueryParam("arch", "all")
	res, err = h.NodeCheck(both)
	require.NoError(t, err)
	require.Len(t, res.Nodes, 2)
	assert.Equal(t, []string{"x86_64", "aarch64"}, []string{res.Nodes[0].Arch, res.Nodes[1].Arch})
	assert.Equal(t, []bool{true, false}, []bool{res.Nodes[0].Ok, res.Nodes[1].Ok})

	// the kernel path is templated on the arch, so another arch is another file
	c := renderRequest("")
	c.SetQueryParam("arch", "aarch64")
	res, err = h.NodeCheck(c)
	require.NoError(t, err)
	assert.False(t, res.Nodes[0].Ok)
	assert.Equal(t, filepath.Join(dir, "vmlinuz-aarch64"), res.Nodes[0].Files[0].Path)
}

func TestReadableFile(t *testing.T) {
	dir := t.TempDir()
	_, err := readableFile(dir)
	assert.ErrorContains(t, err, "not a regular file")

	path := filepath.Join(dir, "vmlinuz")
	require.NoError(t, os.WriteFile(path, []byte("kernel"), 0o644))
	size, err := readableFile(path)
	require.NoError(t, err)
	assert.EqualValues(t, 6, size)
}
