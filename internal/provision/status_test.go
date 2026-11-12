// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/beacon"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

// httptest.NewRequest comes from 192.0.2.1, so the host owns that address unless a test says otherwise
func statusHost(t *testing.T, h *Handler, provision bool) *model.Host {
	t.Helper()

	image := &model.BootImage{Name: "flatcar", KernelPath: "/vmlinuz"}
	require.NoError(t, h.DB.StoreBootImage(image))

	host := &model.Host{
		Name:      "cpn-01",
		BootImage: image.Name,
		Provision: provision,
		Interfaces: []*model.NetInterface{{
			MAC: net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			IP:  netip.MustParsePrefix("192.0.2.1/24"),
		}},
	}
	require.NoError(t, h.DB.StoreHost(host))
	return host
}

func postStatus(t *testing.T, h *Handler, host *model.Host, body string) (*httptest.ResponseRecorder, error) {
	t.Helper()

	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	require.NoError(t, err)

	e := newTestEcho(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c := e.NewContext(req, rec)
	c.SetPath("/boot/:token/status")
	c.SetParamNames("token")
	c.SetParamValues(token)

	return rec, TokenRequired(h.Status)(c)
}

func httpCode(t *testing.T, err error) int {
	t.Helper()
	var he *echo.HTTPError
	require.True(t, errors.As(err, &he), "expected an echo.HTTPError, got %v", err)
	return he.Code
}

func TestStatusReport(t *testing.T) {
	t.Run("stored without the node set to provision", func(t *testing.T) {
		h := &Handler{DB: newTestDB(t)}
		host := statusHost(t, h, false)

		rec, err := postStatus(t, h, host, `{"state":"failed","message":"1 failed","boot_id":"b1","failed_units":[{"name":"raid.service","result":"exit-code"}]}`)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		s, err := h.DB.LoadProvisionStatus(host.UID.String())
		require.NoError(t, err)
		assert.Equal(t, model.ProvisionStateFailed, s.State)
		assert.Equal(t, "1 failed", s.Message)
		assert.Equal(t, "192.0.2.1", s.Addr)
		assert.Equal(t, []model.FailedUnit{{Name: "raid.service", Result: "exit-code"}}, s.FailedUnits)
		assert.Nil(t, s.CompletedAt)
	})

	t.Run("refused from another address", func(t *testing.T) {
		h := &Handler{DB: newTestDB(t)}
		host := statusHost(t, h, true)
		host.Interfaces[0].IP = netip.MustParsePrefix("10.0.0.5/24")
		require.NoError(t, h.DB.StoreHost(host))

		_, err := postStatus(t, h, host, `{"state":"completed"}`)
		assert.Equal(t, http.StatusForbidden, httpCode(t, err))

		_, err = h.DB.LoadProvisionStatus(host.UID.String())
		assert.ErrorIs(t, err, store.ErrNotFound)

		viper.Set("provision.status_check_source", false)
		t.Cleanup(func() { viper.Set("provision.status_check_source", true) })
		_, err = postStatus(t, h, host, `{"state":"completed"}`)
		assert.NoError(t, err)
	})

	t.Run("unknown state", func(t *testing.T) {
		h := &Handler{DB: newTestDB(t)}
		host := statusHost(t, h, true)

		_, err := postStatus(t, h, host, `{"state":"pxe"}`)
		assert.Equal(t, http.StatusBadRequest, httpCode(t, err))
	})

	t.Run("complete unprovisions only when asked", func(t *testing.T) {
		h := &Handler{DB: newTestDB(t)}
		host := statusHost(t, h, true)

		_, err := postStatus(t, h, host, `{"state":"completed","boot_id":"b1"}`)
		require.NoError(t, err)
		loaded, err := h.DB.LoadHostFromID(host.UID.String())
		require.NoError(t, err)
		assert.True(t, loaded.Provision)

		_, err = postStatus(t, h, host, `{"state":"completed","boot_id":"b1","complete":true}`)
		require.NoError(t, err)
		loaded, err = h.DB.LoadHostFromID(host.UID.String())
		require.NoError(t, err)
		assert.False(t, loaded.Provision)

		s, err := h.DB.LoadProvisionStatus(host.UID.String())
		require.NoError(t, err)
		assert.Equal(t, model.ProvisionStateCompleted, s.State)
		assert.NotNil(t, s.CompletedAt)
	})

	t.Run("finished provision only hears from its own boot", func(t *testing.T) {
		h := &Handler{DB: newTestDB(t)}
		host := statusHost(t, h, true)

		_, err := postStatus(t, h, host, `{"state":"completed","boot_id":"b1"}`)
		require.NoError(t, err)

		_, err = postStatus(t, h, host, `{"state":"failed","boot_id":"b2"}`)
		assert.Equal(t, http.StatusConflict, httpCode(t, err))
		_, err = postStatus(t, h, host, `{"state":"failed"}`)
		assert.Equal(t, http.StatusConflict, httpCode(t, err))

		_, err = postStatus(t, h, host, `{"state":"failed","boot_id":"b1"}`)
		assert.NoError(t, err)
	})

	t.Run("installer report without a boot id lets the installed os report", func(t *testing.T) {
		h := &Handler{DB: newTestDB(t)}
		host := statusHost(t, h, true)

		_, err := postStatus(t, h, host, `{"state":"completed"}`)
		require.NoError(t, err)
		_, err = postStatus(t, h, host, `{"state":"completed","boot_id":"os"}`)
		assert.NoError(t, err)
	})

	t.Run("oversized body", func(t *testing.T) {
		h := &Handler{DB: newTestDB(t)}
		host := statusHost(t, h, true)

		_, err := postStatus(t, h, host, `{"state":"failed","message":"`+strings.Repeat("x", maxStatusBody)+`"}`)
		assert.Equal(t, http.StatusBadRequest, httpCode(t, err))
	})
}

func TestStatusStages(t *testing.T) {
	h := &Handler{DB: newTestDB(t)}
	host := statusHost(t, h, true)
	uid := host.UID.String()

	load := func() *model.ProvisionStatus {
		t.Helper()
		s, err := h.DB.LoadProvisionStatus(uid)
		require.NoError(t, err)
		return s
	}

	RecordPXE(h.DB, host)
	first := load()
	assert.Equal(t, model.ProvisionStatePXE, first.State)

	// iPXE chainloading asks DHCP again in the same boot
	RecordPXE(h.DB, host)
	assert.Equal(t, first.StartedAt, load().StartedAt)

	h.recordStage(host, model.ProvisionStateIPXE, nil)
	s := load()
	assert.Equal(t, model.ProvisionStateIPXE, s.State)
	assert.True(t, s.Stages[model.ProvisionStatePXE].Equal(first.StartedAt), "the first stage is when the provision started")
	assert.Contains(t, s.Stages, model.ProvisionStateIPXE)
	ipxeAt := s.Stages[model.ProvisionStateIPXE]

	// a repeated stage keeps the time it was first reached
	h.recordStage(host, model.ProvisionStateIPXE, nil)
	assert.True(t, load().Stages[model.ProvisionStateIPXE].Equal(ipxeAt))

	_, err := postStatus(t, h, host, `{"state":"failed","boot_id":"b1","failed_units":[{"name":"x.service"}]}`)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	RecordPXE(h.DB, host)
	s = load()
	assert.Equal(t, model.ProvisionStatePXE, s.State)
	assert.Len(t, s.Stages, 1, "a new attempt drops the last one's stage times")
	assert.True(t, s.StartedAt.After(first.StartedAt), "a network boot after a finished provision starts a new one")
	assert.Empty(t, s.FailedUnits)
	assert.Empty(t, s.BootID)

	h.recordCompleted(host)
	s = load()
	assert.Equal(t, model.ProvisionStateCompleted, s.State)
	require.NotNil(t, s.CompletedAt)

	ns, err := nodeset.NewNodeSet(host.Name)
	require.NoError(t, err)
	list, err := h.DB.ProvisionStatuses(ns)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, host.Name, list[0].Node)

	// the status goes with the node
	require.NoError(t, h.DB.DeleteHosts(ns))
	list, err = h.DB.ProvisionStatuses(nil)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestCompleteRecordsStatus(t *testing.T) {
	complete := func(t *testing.T, h *Handler, host *model.Host, query string) error {
		token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
		require.NoError(t, err)

		e := newTestEcho(t)
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/"+query, nil), httptest.NewRecorder())
		c.SetPath("/boot/:token/complete")
		c.SetParamNames("token")
		c.SetParamValues(token)
		return TokenRequired(h.Complete)(c)
	}

	for _, tc := range []struct {
		query         string
		wantProvision bool
	}{
		{"", false},
		{"?unprovision=true", false},
		{"?unprovision=false", true},
	} {
		t.Run("query "+tc.query, func(t *testing.T) {
			h := &Handler{DB: newTestDB(t)}
			host := statusHost(t, h, true)
			require.NoError(t, complete(t, h, host, tc.query))

			s, err := h.DB.LoadProvisionStatus(host.UID.String())
			require.NoError(t, err)
			assert.Equal(t, model.ProvisionStateCompleted, s.State)
			assert.NotNil(t, s.CompletedAt)

			got, err := h.DB.LoadHostFromID(host.UID.String())
			require.NoError(t, err)
			assert.Equal(t, tc.wantProvision, got.Provision)
		})
	}

	t.Run("bad value", func(t *testing.T) {
		h := &Handler{DB: newTestDB(t)}
		host := statusHost(t, h, true)
		var he *echo.HTTPError
		require.ErrorAs(t, complete(t, h, host, "?unprovision=maybe"), &he)
		assert.Equal(t, http.StatusBadRequest, he.Code)

		got, err := h.DB.LoadHostFromID(host.UID.String())
		require.NoError(t, err)
		assert.True(t, got.Provision, "a rejected request changes nothing")
		_, err = h.DB.LoadProvisionStatus(host.UID.String())
		assert.Error(t, err, "no status is recorded")
	})
}

func TestBeaconDownload(t *testing.T) {
	h := &Handler{}
	get := func(arch string) (*httptest.ResponseRecorder, error) {
		e := newTestEcho(t)
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
		c.SetPath("/beacon/:arch")
		c.SetParamNames("arch")
		c.SetParamValues(arch)
		return rec, h.Beacon(c)
	}

	_, err := get("amd64")
	assert.Equal(t, http.StatusNotFound, httpCode(t, err), "nothing is served without a dir")

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "grendel-beacon-amd64"), []byte("binary"), 0o755))
	viper.Set("provision.beacon.dir", dir)
	t.Cleanup(func() { viper.Set("provision.beacon.dir", "") })

	rec, err := get("amd64")
	require.NoError(t, err)
	assert.Equal(t, "binary", rec.Body.String())

	_, err = get("arm64")
	assert.Equal(t, http.StatusNotFound, httpCode(t, err))
	_, err = get("..")
	assert.Equal(t, http.StatusNotFound, httpCode(t, err))
}

func TestIgnitionBeacon(t *testing.T) {
	viper.Set("provision.beacon.enabled", true)
	viper.Set("provision.beacon.dir", "/srv/beacon")
	viper.Set("provision.beacon.complete", true)
	t.Cleanup(func() {
		viper.Set("provision.beacon.enabled", false)
		viper.Set("provision.beacon.dir", "")
		viper.Set("provision.beacon.complete", false)
	})

	rec, err := requestIgnition(t, "")
	require.NoError(t, err)

	var ign struct {
		Storage struct {
			Files []struct {
				Path     string `json:"path"`
				Contents struct {
					Source string `json:"source"`
				} `json:"contents"`
			} `json:"files"`
		} `json:"storage"`
		Systemd struct {
			Units []struct {
				Name     string `json:"name"`
				Enabled  bool   `json:"enabled"`
				Contents string `json:"contents"`
			} `json:"units"`
		} `json:"systemd"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &ign))

	files := map[string]string{}
	for _, f := range ign.Storage.Files {
		files[f.Path] = f.Contents.Source
	}
	require.Contains(t, files, "/opt/bin/grendel-beacon")
	assert.Contains(t, files["/opt/bin/grendel-beacon"], "/beacon/")
	require.Contains(t, files, "/etc/grendel/beacon.env")
	assert.Contains(t, files["/etc/grendel/beacon.env"], "%2Fstatus", "the status url is in the env file")

	require.Len(t, ign.Systemd.Units, 1)
	assert.True(t, ign.Systemd.Units[0].Enabled)
	assert.Contains(t, ign.Systemd.Units[0].Contents, "ExecStart=/opt/bin/grendel-beacon watch --complete")
}

// the beacon keeps its own copy of the report types so it doesn't pull in grendel's server packages, this keeps the two in step
func TestBeaconReportMatchesServer(t *testing.T) {
	in := beacon.Report{
		State:       beacon.StateFailed,
		Message:     "m",
		FailedUnits: []beacon.FailedUnit{{Name: "a", Description: "b", ActiveState: "c", SubState: "d", Result: "e"}},
		QueuedJobs:  3,
		BootID:      "id",
		Complete:    true,
	}
	data, err := json.Marshal(in)
	require.NoError(t, err)

	var out StatusReport
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(&out))

	back, err := json.Marshal(out)
	require.NoError(t, err)
	assert.JSONEq(t, string(data), string(back))
}

func TestStatusFacts(t *testing.T) {
	h := &Handler{DB: newTestDB(t)}
	host := statusHost(t, h, true)

	_, err := postStatus(t, h, host, `{"state":"booting","facts":{"os.id":"rocky","os.image_version":"2026.11.3","Bad Key":"x","kernel.release":"`+strings.Repeat("k", maxFactValue+10)+`"}}`)
	require.NoError(t, err)
	_, err = postStatus(t, h, host, `{"state":"booting","facts":{"os.image_version":"2026.11.4","slurm.version":"24.05"}}`)
	require.NoError(t, err)

	s, err := h.DB.LoadProvisionStatus(host.UID.String())
	require.NoError(t, err)
	assert.Equal(t, "rocky", s.Facts["os.id"], "a later report keeps the facts it doesn't mention")
	assert.Equal(t, "2026.11.4", s.Facts["os.image_version"], "and replaces the ones it does")
	assert.Equal(t, "24.05", s.Facts["slurm.version"])
	assert.NotContains(t, s.Facts, "Bad Key")
	assert.Len(t, s.Facts["kernel.release"], maxFactValue)

	many := map[string]string{}
	for i := range maxFacts + 50 {
		many[fmt.Sprintf("site.f%03d", i)] = "v"
	}
	assert.Len(t, mergeFacts(nil, many), maxFacts)
	assert.Contains(t, mergeFacts(nil, many), "site.f000", "the first keys in sorted order are the ones kept")
}

func TestStatusRecordsImage(t *testing.T) {
	h := &Handler{DB: newTestDB(t)}
	host := statusHost(t, h, true)
	image := &model.BootImage{Name: "frosty", Version: "2026.11.3"}

	h.recordStage(host, model.ProvisionStateIPXE, image)
	h.recordStage(host, model.ProvisionStateConfig, &model.BootImage{Name: "other", Version: "9"})

	s, err := h.DB.LoadProvisionStatus(host.UID.String())
	require.NoError(t, err)
	assert.Equal(t, "frosty", s.BootImage, "the image is the one sent with the iPXE script")
	assert.Equal(t, "2026.11.3", s.BootImageVersion)
}
