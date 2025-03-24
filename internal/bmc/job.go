// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/ubccr/grendel/pkg/model"
)

type Job struct {
	delay  time.Duration
	fanout int
}

func NewJob() *Job {
	return &Job{
		delay:  time.Duration(viper.GetInt("bmc.delay")) * time.Second,
		fanout: viper.GetInt("bmc.fanout"),
	}
}

func PrintStatusCli(output model.JobMessageList) {
	for _, m := range output {
		log.Warnf("Error during redfish query: %s\t %s\t %s", m.Status, m.Host, m.Msg)

	}
}

func FormatOutput(output chan model.JobMessage) (model.JobMessageList, error) {
	arr := model.JobMessageList{}
	for m := range output {
		if m.Status == "error" {
			m.RedfishError = ParseRedfishError(errors.New(m.Msg))
		}
		arr = append(arr, m)
	}

	return arr, nil
}

func (j *Job) PowerControl(hostList model.HostList, bootOption redfish.BootSourceOverrideTarget, powerOption redfish.ResetType) (model.JobMessageList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunPowerControl(host, ch, bootOption, powerOption)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) BmcStatus(hostList model.HostList) ([]model.RedfishSystem, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunBmcStatus(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	arr := []model.RedfishSystem{}
	for m := range ch {
		if m.Status != "success" {
			log.Warnf("Error during redfish query: %s\t %s\t %s", m.Status, m.Host, m.Msg)
			continue
		}
		d := model.RedfishSystem{}
		err := json.Unmarshal([]byte(m.Msg), &d)
		if err != nil {
			return nil, err
		}

		arr = append(arr, d)
	}

	return arr, nil
}

func (j *Job) GetFirmware(hostList model.HostList) ([]Firmware, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunGetFirmware(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	arr := []Firmware{}
	for m := range ch {
		if m.Status != "success" {
			log.Warnf("Error during redfish query: %s\t %s\t %s", m.Status, m.Host, m.Msg)
			continue
		}
		d := Firmware{}
		err := json.Unmarshal([]byte(m.Msg), &d)
		if err != nil {
			return nil, err
		}

		arr = append(arr, d)
	}

	return arr, nil
}

func (j *Job) UpdateFirmware(hostList model.HostList, firmwarePaths []string) (model.JobMessageList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunUpdateFirmware(host, ch, firmwarePaths)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)
}

func (j *Job) GetJobs(hostList model.HostList) (model.RedfishJobList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunGetJobs(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	arr := model.RedfishJobList{}
	for m := range ch {
		if m.Status != "success" {
			log.Warnf("Error during redfish query: %s\t %s\t %s", m.Status, m.Host, m.Msg)
			continue
		}
		d := model.RedfishJob{}
		err := json.Unmarshal([]byte(m.Msg), &d)
		if err != nil {
			return nil, err
		}

		layout := "2006-01-02T15:04:05-07:00"
		sort.Slice(d.Jobs, func(i, j int) bool {
			ti, err := time.Parse(layout, d.Jobs[i].StartTime)
			if err != nil {
				return false
			}
			tj, err := time.Parse(layout, d.Jobs[j].StartTime)
			if err != nil {
				return false
			}

			return tj.Before(ti)
		})

		arr = append(arr, d)
	}

	sort.Slice(arr, func(i, j int) bool { return arr[i].Host > arr[j].Host })

	return arr, nil
}

func (j *Job) ClearJobs(hostList model.HostList, ids []string) (model.JobMessageList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunClearJobs(host, ch, ids)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) PowerCycleBmc(hostList model.HostList) (model.JobMessageList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunPowerCycleBmc(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) ClearSel(hostList model.HostList) (model.JobMessageList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunClearSel(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) BmcAutoConfigure(hostList model.HostList) (model.JobMessageList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunBmcAutoConfigure(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)
}

func (j *Job) BmcImportConfiguration(hostList model.HostList, shutdownType, file string) (model.JobMessageList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunBmcImportConfiguration(host, ch, shutdownType, file)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) BmcGetMetricReports(hostList model.HostList) (model.RedfishMetricReportList, error) {
	runner := newJobRunner(j)

	ch := make(chan model.JobMessage, len(hostList))
	for i, host := range hostList {
		if host.HostType() != "server" {
			continue
		}
		runner.RunBmcGetMetricReports(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	arr := model.RedfishMetricReportList{}
	for m := range ch {
		if m.Status != "success" {
			log.Warnf("Error during redfish query: %s\t %s\t %s", m.Status, m.Host, m.Msg)
			continue
		}

		d := model.RedfishMetricReport{Name: m.Host}
		err := json.Unmarshal([]byte(m.Msg), &d.Reports)
		if err != nil {
			return nil, err
		}

		arr = append(arr, d)
	}

	return arr, nil
}
