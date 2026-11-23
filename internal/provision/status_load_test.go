// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/events"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/pkg/model"
)

// TestProvisionWaveLoad drives a whole cluster rebooting at once through the provision status path, against a database file as grendeld runs it, and reports how long each report took. It is slow, so it only runs when GRENDEL_LOAD_NODES is set, to the number of nodes:
//
//	GRENDEL_LOAD_NODES=1000 go test ./internal/provision -run TestProvisionWaveLoad -v
func TestProvisionWaveLoad(t *testing.T) {
	nodes, _ := strconv.Atoi(os.Getenv("GRENDEL_LOAD_NODES"))
	if nodes <= 0 {
		t.Skip("set GRENDEL_LOAD_NODES to run")
	}
	const workers = 64
	const beaconReports = 3

	// the database goes in GRENDEL_LOAD_DIR when set, as a temporary directory may be in memory, where writes cost nothing to make durable
	dir := os.Getenv("GRENDEL_LOAD_DIR")
	if dir == "" {
		dir = t.TempDir()
	} else {
		made, err := os.MkdirTemp(dir, "grendel-load-")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(made) })
		dir = made
	}
	db, err := sqlstore.New(filepath.Join(dir, "grendel.db"))
	require.NoError(t, err)
	defer db.Close()
	h := &Handler{DB: db}

	image := &model.BootImage{Name: "flatcar", KernelPath: "/vmlinuz"}
	require.NoError(t, db.StoreBootImage(image))
	hosts := make([]*model.Host, nodes)
	for i := range hosts {
		hosts[i] = &model.Host{
			Name:      fmt.Sprintf("cpn-%05d", i),
			BootImage: image.Name,
			Provision: true,
			Interfaces: []*model.NetInterface{{
				MAC: net.HardwareAddr{0x02, 0, 0, byte(i >> 16), byte(i >> 8), byte(i)},
				IP:  netip.PrefixFrom(netip.AddrFrom4([4]byte{10, byte(i >> 16), byte(i >> 8), byte(i)}), 8),
			}},
		}
		require.NoError(t, db.StoreHost(hosts[i]))
	}

	// each node's reports, in the order a node sends them
	steps := []func(*model.Host){
		func(host *model.Host) { RecordPXE(db, host) },
		func(host *model.Host) { h.recordStage(host, model.ProvisionStateIPXE, image) },
		func(host *model.Host) { h.recordStage(host, model.ProvisionStateConfig, nil) },
	}
	for range beaconReports {
		steps = append(steps, func(host *model.Host) {
			recordStatus(db, host, func(s *model.ProvisionStatus, now time.Time) { s.State = model.ProvisionStateBooting })
		})
	}
	steps = append(steps, h.recordCompleted)

	// the nodes move through their steps together, as a rack powered on at once does, with reports from different nodes at the same time
	// the recorder tidies as grendeld runs it, competing with the reports for the database
	ctx, cancel := context.WithCancel(context.Background())
	recorded := make(chan error, 1)
	go func() { recorded <- events.New(db).Run(ctx) }()

	var mu sync.Mutex
	var took []time.Duration
	started := time.Now()
	for _, step := range steps {
		work := make(chan *model.Host)
		var wg sync.WaitGroup
		for range workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var mine []time.Duration
				for host := range work {
					t0 := time.Now()
					step(host)
					mine = append(mine, time.Since(t0))
				}
				mu.Lock()
				took = append(took, mine...)
				mu.Unlock()
			}()
		}
		for _, host := range hosts {
			work <- host
		}
		close(work)
		wg.Wait()
	}
	wall := time.Since(started)
	cancel()
	require.NoError(t, <-recorded)

	slices.Sort(took)
	pct := func(p float64) time.Duration { return took[int(float64(len(took)-1)*p)] }
	t.Logf("%d nodes, %d reports in %s (%.0f reports/s)", nodes, len(took), wall.Round(time.Millisecond), float64(len(took))/wall.Seconds())
	t.Logf("per report: p50 %s  p90 %s  p99 %s  max %s", pct(.5), pct(.9), pct(.99), took[len(took)-1])

	list, err := db.Events(model.EventFilter{Kind: model.EventKindProvision})
	require.NoError(t, err)
	require.Len(t, list, 1, "every node joins the one provisioning")
	require.Equal(t, model.EventStateDone, list[0].State)
	require.Equal(t, model.EventCounts{Total: nodes, Succeeded: nodes}, list[0].Counts)
}
