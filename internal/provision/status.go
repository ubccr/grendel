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
	"net/netip"
	"path/filepath"
	"sort"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/beacon/wire"
	"github.com/ubccr/grendel/internal/events"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

const (
	// pxeRestartAfter is how long a node can sit in the pxe state before another PXE request counts as a new attempt rather than the next step of iPXE chainloading
	pxeRestartAfter = 10 * time.Minute

	maxStatusBody   = 512 << 10
	maxFailedUnits  = 100
	maxStatusString = 512
	maxStatusMsg    = 2048
	maxFacts        = 64
	maxFactValue    = 1024
)

var errOtherBoot = errors.New("provision already finished in another boot")

func init() {
	viper.SetDefault("provision.status_check_source", true)
	viper.SetDefault("provision.beacon.enabled", false)
	viper.SetDefault("provision.beacon.dir", "")
	viper.SetDefault("provision.beacon.path", "/opt/bin/grendel-beacon")
	viper.SetDefault("provision.beacon.complete", false)
}

// recordStatus changes the host's provision status with update. Tracking status must never get in the way of a boot, so a failure is only logged
func recordStatus(db store.Store, host *model.Host, update func(s *model.ProvisionStatus, now time.Time)) {
	err := updateStatus(db, host, func(s *model.ProvisionStatus, now time.Time) error {
		update(s, now)
		return nil
	})
	if err != nil {
		log.Warnf("failed to record provision status for %s: %s", host.Name, err)
	}
}

// updateStatus lets update change the host's provision status and writes it with the node's row in the provisioning event, in one transaction. An error from update is returned with nothing written
func updateStatus(db store.Store, host *model.Host, update func(s *model.ProvisionStatus, now time.Time) error) error {
	now := time.Now().UTC()

	return db.UpdateProvisionStatus(host.UID.String(), events.Provisioning(now), func(status *model.ProvisionStatus) (*model.EventNode, error) {
		if status.StartedAt.IsZero() {
			status.StartedAt = now
		}
		if err := update(status, now); err != nil {
			return nil, err
		}
		status.UpdatedAt = now
		if status.State == "" {
			return nil, nil
		}

		if status.Stages == nil {
			status.Stages = map[model.ProvisionState]time.Time{}
		}
		if _, ok := status.Stages[status.State]; !ok {
			status.Stages[status.State] = now
		}

		return eventNode(host.Name, status), nil
	})
}

// eventNode is the node's row in the provisioning event. Every report writes the row, the same state again too, since a node that goes quiet is what counts as stalled
func eventNode(node string, status *model.ProvisionStatus) *model.EventNode {
	// the node's message, such as why its boot failed, is carried onto its row so the events page can show it
	row := &model.EventNode{Node: node, State: model.EventNodeRunning, Stage: string(status.State), Message: status.Message, UpdatedAt: status.UpdatedAt}
	switch status.State {
	case model.ProvisionStateCompleted:
		row.State = model.EventNodeSucceeded
	case model.ProvisionStateFailed:
		row.State = model.EventNodeFailed
	}
	return row
}

// restart starts a new provision attempt, dropping everything the last one reported
func restart(s *model.ProvisionStatus, state model.ProvisionState, now time.Time) {
	*s = model.ProvisionStatus{State: state, StartedAt: now}
}

// RecordPXE marks the start of a provision when a node asks DHCP for a network boot. iPXE chainloading sends several PXE requests in one boot, so a node already in the pxe state stays in the same attempt
func RecordPXE(db store.Store, host *model.Host) {
	recordStatus(db, host, func(s *model.ProvisionStatus, now time.Time) {
		if s.State == model.ProvisionStatePXE && now.Sub(s.UpdatedAt) < pxeRestartAfter {
			return
		}
		restart(s, model.ProvisionStatePXE, now)
	})
}

// recordStage moves a node on to state when it fetches its iPXE script or config. A node whose last provision finished is starting a new one, which happens when DHCP is served by something other than grendel. The image is recorded when the node fetches its iPXE script, since that is the image it was sent
func (h *Handler) recordStage(host *model.Host, state model.ProvisionState, image *model.BootImage) {
	recordStatus(h.DB, host, func(s *model.ProvisionStatus, now time.Time) {
		if s.State == "" || s.State.Finished() {
			restart(s, state, now)
		} else {
			s.State = state
		}
		if image != nil && (state == model.ProvisionStateIPXE || s.BootImage == "") {
			s.BootImage = image.Name
			s.BootImageVersion = image.Version
		}
	})
}

// mergeFacts adds the valid facts in report to current, keeping at most maxFacts. Keys are taken in sorted order so which ones are dropped over the limit doesn't change from one report to the next
func mergeFacts(current, report map[string]string) map[string]string {
	if len(report) == 0 {
		return current
	}
	merged := make(map[string]string, len(current)+len(report))
	for k, v := range current {
		merged[k] = v
	}

	keys := make([]string, 0, len(report))
	for k := range report {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if !wire.ValidFactKey(k) {
			continue
		}
		if _, ok := merged[k]; !ok && len(merged) >= maxFacts {
			continue
		}
		merged[k] = truncate(report[k], maxFactValue)
	}
	return merged
}

func (h *Handler) recordCompleted(host *model.Host) {
	recordStatus(h.DB, host, func(s *model.ProvisionStatus, now time.Time) {
		if s.State == "" {
			s.StartedAt = now
		}
		s.State = model.ProvisionStateCompleted
		if s.CompletedAt == nil {
			s.CompletedAt = &now
		}
	})
}

// unprovision turns provisioning off for host, flipping only the provision flag since saving the whole host would persist the arch detected for this boot
func (h *Handler) unprovision(host *model.Host) error {
	log.Infof("Unprovisioning host %s", host.Name)

	ns, err := nodeset.NewNodeSet(host.Name)
	if err == nil {
		err = h.DB.ProvisionHosts(ns, false)
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to unprovision host").SetInternal(err)
	}

	return nil
}

// Status takes a report from the beacon running in the node's OS. Unlike the other boot routes it does not need the node to still be set to provision, so the final report lands after something else has completed it
func (h *Handler) Status(c echo.Context) error {
	claims := c.Get(ContextKeyToken).(*model.BootClaims)

	host, err := h.DB.LoadHostFromID(claims.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid host").SetInternal(err)
	}

	addr, err := peerAddr(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid remote address").SetInternal(err)
	}

	// the token alone would let anyone who saw it report for the node, so the report also has to come from one of the node's own addresses. This is the TCP peer, since X-Forwarded-For is whatever the client says
	if viper.GetBool("provision.status_check_source") && !hostHasAddr(host, addr) {
		return echo.NewHTTPError(http.StatusForbidden, "status must come from one of the node's addresses").SetInternal(fmt.Errorf("report for %s came from %s", host.Name, addr))
	}

	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxStatusBody)
	var report wire.Report
	if err := json.NewDecoder(c.Request().Body).Decode(&report); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status report").SetInternal(err)
	}

	state := model.ProvisionState(report.State)
	switch state {
	case model.ProvisionStateBooting, model.ProvisionStateFailed, model.ProvisionStateCompleted:
	default:
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("state must be one of booting, failed, completed, got %q", report.State))
	}

	log.Infof("Provision status from host %s: %s", host.Name, state)

	err = updateStatus(h.DB, host, func(s *model.ProvisionStatus, now time.Time) error {
		// a finished provision only takes more reports from the boot that finished it, so a token copied off the node can't reopen it
		if s.State.Finished() && s.BootID != "" && report.BootID != s.BootID {
			return errOtherBoot
		}

		s.State = state
		s.Message = truncate(report.Message, maxStatusMsg)
		s.FailedUnits = trimUnits(report.FailedUnits)
		s.QueuedJobs = max(report.QueuedJobs, 0)
		s.BootID = truncate(report.BootID, maxStatusString)
		s.Addr = addr.String()
		s.Facts = mergeFacts(s.Facts, report.Facts)
		if state == model.ProvisionStateCompleted {
			if s.CompletedAt == nil {
				s.CompletedAt = &now
			}
		} else {
			s.CompletedAt = nil
		}
		return nil
	})
	if errors.Is(err, errOtherBoot) {
		return echo.NewHTTPError(http.StatusConflict, errOtherBoot.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to record provision status").SetInternal(err)
	}

	if report.Complete && state == model.ProvisionStateCompleted && host.Provision {
		if err := h.unprovision(host); err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

// Beacon serves the beacon binary for an architecture from provision.beacon.dir, for templates that download it rather than expect it in the image
func (h *Handler) Beacon(c echo.Context) error {
	dir := viper.GetString("provision.beacon.dir")
	arch := c.Param("arch")
	if dir == "" || (arch != "amd64" && arch != "arm64") {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	return c.File(filepath.Join(dir, "grendel-beacon-"+arch))
}

func peerAddr(r *http.Request) (netip.Addr, error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, err
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, err
	}
	return addr.Unmap(), nil
}

func hostHasAddr(host *model.Host, addr netip.Addr) bool {
	for _, nic := range host.Interfaces {
		if nic.IP.IsValid() && nic.Addr().Unmap() == addr {
			return true
		}
	}
	for _, bond := range host.Bonds {
		if bond.IP.IsValid() && bond.Addr().Unmap() == addr {
			return true
		}
	}
	return false
}

func trimUnits(units []wire.FailedUnit) []model.FailedUnit {
	if len(units) > maxFailedUnits {
		units = units[:maxFailedUnits]
	}
	out := make([]model.FailedUnit, 0, len(units))
	for _, u := range units {
		out = append(out, model.FailedUnit{
			Name:        truncate(u.Name, maxStatusString),
			Description: truncate(u.Description, maxStatusString),
			ActiveState: truncate(u.ActiveState, maxStatusString),
			SubState:    truncate(u.SubState, maxStatusString),
			Result:      truncate(u.Result, maxStatusString),
		})
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
