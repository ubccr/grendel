// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"fmt"

	"github.com/korovkin/limiter"
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/ubccr/grendel/pkg/model"
)

type jobRunner struct {
	limit    *limiter.ConcurrencyLimiter
	user     string
	pass     string
	insecure bool
}

func newJobRunner(j *Job) *jobRunner {
	user := viper.GetString("bmc.user")
	pass := viper.GetString("bmc.password")
	insecure := viper.GetBool("bmc.insecure")

	return &jobRunner{
		limit:    limiter.NewConcurrencyLimiter(j.fanout),
		user:     user,
		pass:     pass,
		insecure: insecure,
	}
}

func (r *jobRunner) Wait() {
	r.limit.Wait()
}

func (r *jobRunner) RunPowerControl(host *model.Host, ch chan model.JobMessage, bootOverride redfish.BootSourceOverrideTarget, powerOption redfish.ResetType) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.PowerControl(powerOption, bootOverride)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = "Sent power command"
	})
}

func (r *jobRunner) RunBmcStatus(host *model.Host, ch chan model.JobMessage) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		data := &model.RedfishSystem{}
		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		data, err = r.GetSystem()
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		data.Name = host.Name
		output, err := json.Marshal(data)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = string(output)
	})
}

func (r *jobRunner) RunGetJobs(host *model.Host, ch chan model.JobMessage) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		data := model.RedfishJob{}
		data.Jobs, err = r.GetJobs()
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		data.Host = host.Name

		output, err := json.Marshal(data)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = string(output)
	})
}

func (r *jobRunner) RunClearJobs(host *model.Host, ch chan model.JobMessage, ids []string) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.ClearJobs(ids)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = "Cleared Jobs from the BMC"
	})
}

func (r *jobRunner) RunPowerCycleBmc(host *model.Host, ch chan model.JobMessage) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.PowerCycleBmc()
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}
		m.Status = "success"
		m.Msg = "Sent power cycle bmc command"
	})
}

func (r *jobRunner) RunClearSel(host *model.Host, ch chan model.JobMessage) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.ClearSel()
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}
		m.Status = "success"
		m.Msg = "Sent clear sel command"
	})
}

func (r *jobRunner) RunBmcAutoConfigure(host *model.Host, ch chan model.JobMessage) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.BmcAutoConfigure()
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}
		m.Status = "success"
		m.Msg = "Set AutoConfig to Enabled Once"
	})
}

func (r *jobRunner) RunBmcImportConfiguration(host *model.Host, ch chan model.JobMessage, shutdownType, file string) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		mac := ""
		ip := ""
		if bmc != nil {
			mac = bmc.MAC.String()
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		token, err := model.NewBootToken(host.UID.String(), mac)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
		}
		path := fmt.Sprintf("/boot/%s/bmc", token)

		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		jid, err := r.BmcImportConfiguration(shutdownType, path, file)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = "Submitted import Configuration job:" + jid
	})
}

func (r *jobRunner) RunBmcGetMetricReports(host *model.Host, ch chan model.JobMessage) {
	r.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}

		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = err.Error()
			return
		}

		defer r.client.Logout()

		reports, err := r.BmcGetMetricReports()
		if err != nil {
			m.Msg = err.Error()
			return
		}

		output, err := json.Marshal(reports)
		if err != nil {
			m.Msg = err.Error()
			return
		}

		m.Status = "success"
		m.Msg = string(output)
	})
}

func (jr *jobRunner) RunDellInstallFromRepo(host *model.Host, ch chan model.JobMessage, installBody dell.InstallFromRepoBody) {
	jr.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, jr.user, jr.pass, jr.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		res, err := r.DellInstallFromRepo(installBody)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}
		output, err := json.Marshal(res)
		if err != nil {
			m.Msg = err.Error()
			return
		}

		m.Status = "success"
		m.Msg = "successfully queued firmware update job"
		m.Data = string(output)
	})
}

func (jr *jobRunner) RunDellGetRepoUpdateList(host *model.Host, ch chan model.JobMessage) {
	jr.limit.Execute(func() {
		m := model.JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		} else {
			m.Msg = "failed to find bmc interface to query"
			return
		}
		r, err := NewRedfishClient(ip, jr.user, jr.pass, jr.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		ul, err := r.DellGetRepoUpdateList()
		if err != nil {
			m.Msg = err.Error()
			return
		}

		output, err := json.Marshal(ul)
		if err != nil {
			m.Msg = err.Error()
			return
		}

		m.Status = "success"
		m.Msg = string(output)
	})
}
