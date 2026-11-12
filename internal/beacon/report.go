// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package beacon reports a node's provision status from its OS back to grendel. It deliberately imports nothing else from grendel, so the binary built from it stays small
package beacon

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// States the beacon reports, matching model.ProvisionState
const (
	StateBooting   = "booting"
	StateFailed    = "failed"
	StateCompleted = "completed"
)

// FailedUnit matches model.FailedUnit
type FailedUnit struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ActiveState string `json:"active_state,omitempty"`
	SubState    string `json:"sub_state,omitempty"`
	Result      string `json:"result,omitempty"`
}

// Report matches provision.StatusReport, the body of POST /boot/:token/status
type Report struct {
	State       string            `json:"state"`
	Message     string            `json:"message,omitempty"`
	FailedUnits []FailedUnit      `json:"failed_units,omitempty"`
	QueuedJobs  int               `json:"queued_jobs,omitempty"`
	BootID      string            `json:"boot_id,omitempty"`
	Complete    bool              `json:"complete,omitempty"`
	Facts       map[string]string `json:"facts,omitempty"`
}

// ErrRejected means grendel refused the report, most often because the boot token expired. Sending it again won't help
var ErrRejected = errors.New("grendel rejected the report")

// Client posts reports to a node's status URL
type Client struct {
	URL  string
	HTTP *http.Client
}

func NewClient(url string, insecure bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{URL: url, HTTP: &http.Client{Timeout: 15 * time.Second, Transport: transport}}
}

// Post sends one report. A 4xx answer wraps ErrRejected, anything else that fails is worth retrying
func (c *Client) Post(ctx context.Context, r Report) error {
	body, err := json.Marshal(r)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRejected, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		return nil
	}

	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	err = fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(msg)))
	if resp.StatusCode < 500 {
		return fmt.Errorf("%w: %w", ErrRejected, err)
	}
	return err
}

// PostRetry sends a report until it is accepted, rejected, ctx ends, or giveUp has passed, backing off between attempts
func (c *Client) PostRetry(ctx context.Context, r Report, giveUp time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, giveUp)
	defer cancel()

	wait := time.Second
	for {
		err := c.Post(ctx, r)
		if err == nil || errors.Is(err, ErrRejected) {
			return err
		}
		logf("report failed, retrying in %s: %s", wait, err)

		select {
		case <-ctx.Done():
			return fmt.Errorf("gave up sending the report: %w", err)
		case <-time.After(wait):
		}
		wait = min(wait*2, 30*time.Second)
	}
}

// BootID identifies the running boot, so grendel can tell a late report from this boot apart from one sent with a copied token
func BootID() string {
	id, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(id))
}

func logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
