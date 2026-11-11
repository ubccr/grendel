// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stmcginnis/gofish/schemas"
)

// firmwarePollInterval is how often a catalog comparison job is checked on. A variable so tests do not wait on it
var firmwarePollInterval = 5 * time.Second

func (s *Service) autoConfigureOne(ctx context.Context, t Target) (*Ack, error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	m, err := dellManager(c)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{"Attributes": map[string]string{"NIC.1.AutoConfig": "Enable Once"}}
	res, err := c.client.Patch(m.ODataID+"/Attributes", payload)
	schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		return nil, err
	}

	return doneWith("set NIC.1.AutoConfig to Enable Once")
}

func (s *Service) importConfigOne(ctx context.Context, t Target, req ImportConfigRequest) (*Ack, error) {
	share, ok := req.ShareNames[t.Name]
	if !ok {
		return nil, errors.New("no share path was built for this node")
	}

	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	m, err := dellManager(c)
	if err != nil {
		return nil, err
	}

	target, err := importConfigTarget(m)
	if err != nil {
		return nil, err
	}

	body := importConfigBody{ImportSystemConfigurationBody: req.Body}
	body.ShareParameters = importShareParameters{ShareParameters: req.Body.ShareParameters, PortNumber: req.PortNumber}
	body.ShareParameters.ShareName = share

	// sent directly rather than through gofish's ImportSystemConfiguration, whose ShareParameters has no PortNumber
	res, err := c.client.Post(target, body)
	defer schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		return nil, err
	}

	location := res.Header.Get("Location")
	if location == "" {
		return nil, errors.New("bmc accepted the import but returned no job location")
	}

	task, err := schemas.GetTask(c.client, location)
	if err != nil {
		return nil, err
	}

	return doneWith("submitted import job %s", task.ID)
}

// importShareParameters adds PortNumber, the share's port. The iDRAC accepts it from iDRAC9 7.00.00, so it is omitted unless set
type importShareParameters struct {
	dell.ShareParameters
	PortNumber string `json:",omitempty"`
}

// importConfigBody shadows the embedded ShareParameters with the one that carries PortNumber
type importConfigBody struct {
	dell.ImportSystemConfigurationBody
	ShareParameters importShareParameters
}

// importConfigTarget reads the Dell OEM import action from the manager, the same way gofish's dell.FromManager does
func importConfigTarget(m *schemas.Manager) (string, error) {
	var raw struct {
		Actions struct {
			Oem struct {
				Import schemas.ActionTarget `json:"#OemManager.ImportSystemConfiguration"`
			}
		}
	}
	if err := json.Unmarshal(m.RawData, &raw); err != nil {
		return "", fmt.Errorf("reading the manager's actions: %w", err)
	}
	if raw.Actions.Oem.Import.Target == "" {
		return "", errors.New("idrac declares no import system configuration action")
	}
	return raw.Actions.Oem.Import.Target, nil
}

func (s *Service) firmwareUpdatesOne(ctx context.Context, t Target) (*FirmwareUpdates, error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	sis, err := dellSoftwareInstallation(c)
	if err != nil {
		return nil, err
	}

	updates, err := sis.GetRepoBasedUpdateList()
	if err != nil {
		return nil, err
	}

	return &FirmwareUpdates{Updates: *updates}, nil
}

// installFirmwareOne asks the iDRAC to compare its firmware against a catalog, and with ApplyUpdate to queue the updates. A comparison alone is waited on, because firmware updates can only list its result once the job finishes
func (s *Service) installFirmwareOne(ctx context.Context, t Target, req InstallFirmwareRequest) (*Ack, error) {
	apply := req.Body.ApplyUpdate == dell.ApplyUpdateTrue

	if req.ClearJobs && apply {
		if _, err := s.clearJobsOne(ctx, t, ClearJobsRequest{All: true}); err != nil {
			return nil, fmt.Errorf("clearing the job queue before the update: %w", err)
		}
	}

	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	sis, err := dellSoftwareInstallation(c)
	if err != nil {
		return nil, err
	}

	body := req.Body
	job, err := sis.InstallFromRepository(&body)
	if err != nil {
		return nil, err
	}

	if apply {
		return doneWith("queued firmware update job %s", job.ID)
	}

	if err := waitForDellJob(ctx, sis.GetClient(), job); err != nil {
		return nil, err
	}

	return doneWith("compared firmware against the catalog in job %s, bmc firmware check lists the available updates", job.ID)
}

// waitForDellJob polls a job until it finishes or ctx ends. A job still running when ctx ends keeps running on the BMC
func waitForDellJob(ctx context.Context, client schemas.Client, job *dell.Job) error {
	if job.ODataID == "" {
		return fmt.Errorf("job %s has no uri to check on", job.ID)
	}

	ticker := time.NewTicker(firmwarePollInterval)
	defer ticker.Stop()

	for {
		switch job.JobState {
		case "Completed":
			return nil
		case "Failed", "CompletedWithErrors":
			return fmt.Errorf("job %s ended %s: %s", job.ID, job.JobState, job.Message.Message)
		}
		if job.PercentComplete >= 100 {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("job %s was still running at %d%% when the node timed out, it continues on the bmc", job.ID, job.PercentComplete)
		case <-ticker.C:
		}

		next, err := dell.GetJob(client, job.ODataID)
		if err != nil {
			return err
		}
		job = next
	}
}

// dellManager returns the iDRAC. The operations that use it are Dell OEM actions, so any other BMC is refused before anything is sent
func dellManager(c *conn) (*schemas.Manager, error) {
	if !c.isDell() {
		return nil, fmt.Errorf("only supported on a Dell iDRAC, this bmc reports vendor %q", c.service.Vendor)
	}

	managers, err := c.service.Managers()
	if err != nil {
		return nil, err
	}
	if len(managers) == 0 {
		return nil, errors.New("bmc exposes no managers")
	}

	return managers[0], nil
}

func dellSoftwareInstallation(c *conn) (*dell.SoftwareInstallationService, error) {
	if !c.isDell() {
		return nil, fmt.Errorf("only supported on a Dell iDRAC, this bmc reports vendor %q", c.service.Vendor)
	}

	systems, err := c.service.Systems()
	if err != nil {
		return nil, err
	}
	if len(systems) == 0 {
		return nil, errors.New("bmc exposes no computer systems")
	}

	ds, err := dell.FromComputerSystem(systems[0])
	if err != nil {
		return nil, err
	}

	return ds.SoftwareInstallationService()
}
