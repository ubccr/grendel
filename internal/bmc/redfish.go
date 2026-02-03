// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/model"
)

// Power will change the hosts power state
func (r *Redfish) PowerControl(resetType schemas.ResetType, bootOverride schemas.BootSource) error {
	ss, err := r.service.Systems()
	if err != nil {
		return err
	}

	err = r.bootOverride(bootOverride)
	if err != nil {
		return err
	}

	for _, s := range ss {
		if s.PowerState == schemas.OffPowerState && resetType == schemas.ForceRestartResetType {
			resetType = schemas.OnResetType
		}

		if _, err := s.Reset(resetType); err != nil {
			return err
		}
	}

	return nil
}

// bootOverride will set the boot override target
func (r *Redfish) bootOverride(bootOption schemas.BootSource) error {
	if bootOption == schemas.NoneBootSource {
		return nil
	}

	boot := schemas.Boot{
		BootSourceOverrideTarget:  bootOption,
		BootSourceOverrideEnabled: schemas.OnceBootSourceOverrideEnabled,
	}

	ss, err := r.service.Systems()
	if err != nil {
		return err
	}

	for _, s := range ss {
		err := s.SetBoot(&boot)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Redfish) GetSystem() (*model.RedfishSystem, error) {
	ss, err := r.service.Systems()
	if err != nil {
		return nil, err
	}

	if len(ss) == 0 {
		return nil, errors.New("failed to find system")
	}

	sys := ss[0]

	dcs, err := dell.FromComputerSystem(ss[0])
	if err != nil {
		return nil, err
	}

	system := &model.RedfishSystem{
		HostName:       sys.HostName,
		BIOSVersion:    sys.BiosVersion,
		SerialNumber:   sys.SKU,
		Manufacturer:   sys.Manufacturer,
		Model:          sys.Model,
		PowerStatus:    string(sys.PowerState),
		Health:         string(sys.Status.Health),
		TotalMemory:    float32(gofish.Deref(sys.MemorySummary.TotalSystemMemoryGiB)),
		ProcessorCount: int(gofish.Deref(sys.ProcessorSummary.LogicalProcessorCount)),
		BootNext:       sys.Boot.BootNext,
		BootOrder:      sys.Boot.BootOrder,
		OEMDell:        dcs.OEMSystem,
	}

	return system, nil
}

func (r *Redfish) GetJobInfo(jid string) (*schemas.Job, error) {
	js, err := r.service.JobService()
	if err != nil {
		return nil, err
	}

	jobs, err := js.Jobs()
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.ID == jid {
			return job, nil
		}
	}

	return nil, fmt.Errorf("unable to find job with JID: %s", jid)
}

func (r *Redfish) GetJobs() ([]*schemas.Job, error) {
	js, err := r.service.JobService()
	if err != nil {
		return nil, err
	}

	return js.Jobs()
}
func (r *Redfish) ClearJobs(ids []string) error {
	js, err := r.service.JobService()
	if err != nil {
		return err
	}

	if len(ids) > 0 && ids[0] == "JID_CLEARALL" {
		jobs, err := js.Jobs()
		if err != nil {
			return err
		}

		for _, job := range jobs {
			uri := fmt.Sprintf("%s/Jobs/%s", js.ODataID, job.ID)
			resp, err := r.client.Delete(uri)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
		}
	} else {
		for _, id := range ids {
			uri := fmt.Sprintf("%s/Jobs/%s", js.ODataID, id) // TODO: find correct odataid
			resp, err := r.client.Delete(uri)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
		}
		return nil
	}

	return nil
}

func (r *Redfish) GetTaskInfo(tid string) (*schemas.Task, error) {
	ts, err := r.service.Tasks()
	if err != nil {
		return nil, err
	}

	tasks, err := ts.Tasks()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.ID == tid {
			return task, nil
		}
	}

	fmt.Printf("%#v\n", tasks)
	return nil, fmt.Errorf("unable to find task with ID: %s", tid)
}

func (r *Redfish) PowerCycleBmc() error {
	ms, err := r.service.Managers()
	if err != nil {
		return err
	}

	for _, m := range ms {
		if _, err := m.Reset(schemas.GracefulRestartResetType); err != nil {
			return err
		}
	}
	return nil
}

func (r *Redfish) ClearSel() error {
	ms, err := r.service.Managers()
	if err != nil {
		return err
	}

	for _, m := range ms {
		ls, err := m.LogServices()
		if err != nil {
			return err
		}
		for _, l := range ls {
			// ClearLog() errors on other logservice types like "FaultList" on Dell...
			// TODO: Find better solution or add more vendor support below 🙄

			// fmt.Printf("\nID: %s\n Type: %s\n Name: %s\n\n", l.ID, l.LogEntryType, l.Name)
			if l.ID == "Sel" || l.ID == "Log1" || len(ls) == 1 {
				if _, err := l.ClearLog(""); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *Redfish) BmcAutoConfigure() error {
	// TODO: Submit PR to gofish to support this natively?

	type attributes struct {
		NIC1AutoConfig string `json:"NIC.1.AutoConfig"`
	}
	type payload struct {
		Attributes attributes
	}

	p := payload{
		Attributes: attributes{
			NIC1AutoConfig: "Enable Once",
		},
	}

	return r.service.Patch("/redfish/v1/Managers/iDRAC.Embedded.1/Attributes", p)
}

func (r *Redfish) BmcImportConfiguration(st, path, file string) (string, error) {
	shareType := dell.HTTPISCShareType

	if viper.IsSet("provision.cert") {
		shareType = dell.HTTPSISCShareType
	}

	shutdownType := dell.NoRebootISCShutdownType

	switch st {
	case "Forced":
		shutdownType = dell.ForcedISCShutdownType
	case "Graceful":
		shutdownType = dell.GracefulISCShutdownType
	case "NoReboot":
		shutdownType = dell.NoRebootISCShutdownType
	}

	rawip, err := util.GetFirstExternalIPFromInterfaces()
	if err != nil {
		return "", err
	}

	ip := rawip.String()
	lip, port, err := net.SplitHostPort(viper.GetString("provision.listen"))
	if err != nil {
		return "", err
	}

	if lip != "0.0.0.0" {
		ip = lip
	}

	cip := viper.GetString("bmc.config_share_ip")
	if cip != "" {
		ip = cip
	}

	icw := dell.DisabledISCIgnoreCertificateWarning
	if viper.GetString("bmc.config_ignore_certificate_warning") == "Enabled" {
		icw = dell.EnabledISCIgnoreCertificateWarning
	}

	log.Debugf("Import system config debug info: scheme=%s ip=%s port=%s file=%s path=%s icw=%s", shareType, ip, port, file, path, icw)

	body := dell.ImportSystemConfigurationBody{
		// ExecutionMode:  dell.DefaultExecutionMode,
		HostPowerState: dell.OnISCHostPowerState,
		ShutdownType:   shutdownType,
		ShareParameters: dell.ShareParameters{
			Target:                   "ALL",
			ShareType:                shareType,
			IPAddress:                ip,
			ShareName:                path,
			FileName:                 file,
			IgnoreCertificateWarning: icw,
		},
	}

	m, err := r.service.Managers()
	if err != nil {
		return "", err
	}

	dm, err := dell.FromManager(m[0])
	if err != nil {
		return "", err
	}

	j, err := dm.ImportSystemConfiguration(&body)
	if err != nil {
		return "", err
	}
	return j.ID, nil
}

func (r *Redfish) BmcGetJob(id string) (*schemas.Job, error) {
	j, err := r.service.JobService()
	if err != nil {
		return nil, err
	}

	jobs, err := j.Jobs()
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.ID == id {
			return job, nil
		}
	}
	return nil, errors.New("failed to find job")
}

func (r *Redfish) BmcGetMetricReports() ([]*schemas.MetricReport, error) {
	ts, err := r.service.TelemetryService()
	if err != nil {
		return nil, err
	}

	mr, err := ts.MetricReports()

	return mr, err
}

func (r *Redfish) DellInstallFromRepo(body dell.InstallFromRepoBody) (*schemas.Job, error) {
	s, err := r.service.Systems()
	if err != nil {
		return nil, err
	}
	js, err := r.service.JobService()
	if err != nil {
		return nil, err
	}

	ds, err := dell.FromComputerSystem(s[0])
	if err != nil {
		return nil, err
	}
	sis, err := ds.SoftwareInstallationService()
	if err != nil {
		return nil, err
	}

	djob, err := sis.InstallFromRepository(&body)
	if err != nil {
		return nil, err
	}
	jid := djob.ID

	jobList, err := js.Jobs()
	if err != nil {
		return nil, err
	}
	var job *schemas.Job
	for _, j := range jobList {
		if j.ID == jid {
			continue
		}

		job = j
		break
	}

	if body.ApplyUpdate == dell.ApplyUpdateTrue {
		return job, nil
	}

	for range 10 {
		jobList, err := js.Jobs()
		if err != nil {
			return nil, err
		}

		for _, j := range jobList {
			if j.ID != jid {
				continue
			}

			job = j
			break
		}

		if gofish.Deref(job.PercentComplete) == 100 {
			break
		}

		time.Sleep(time.Second * 5)
	}

	if gofish.Deref(job.PercentComplete) != 100 {
		return job, errors.New("timed out waiting for update job to complete")
	}

	return job, nil
}

func (r *Redfish) DellGetRepoUpdateList() (*dell.UpdateList, error) {
	s, err := r.service.Systems()
	if err != nil {
		return nil, err
	}

	ds, err := dell.FromComputerSystem(s[0])
	if err != nil {
		return nil, err
	}
	sis, err := ds.SoftwareInstallationService()
	if err != nil {
		return nil, err
	}

	res, err := sis.GetRepoBasedUpdateList()
	return res, err
}
