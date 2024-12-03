// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/viper"
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

func PrintStatusCli(output []JobMessage) {
	for _, m := range output {
		fmt.Printf("%s\t%s\t%s\n", m.Status, m.Host, m.Msg)
	}
}

func FormatOutput(output chan JobMessage) ([]JobMessage, error) {
	arr := []JobMessage{}
	for m := range output {
		if m.Status == "error" {
			m.RedfishError = ParseRedfishError(errors.New(m.Msg))
		}
		arr = append(arr, m)
	}

	return arr, nil
}

func (j *Job) PowerCycle(hostList model.HostList, bootOption string) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
		runner.RunPowerCycle(host, ch, bootOption)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) PowerOn(hostList model.HostList, bootOption string) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
		runner.RunPowerOn(host, ch, bootOption)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) PowerOff(hostList model.HostList) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
		runner.RunPowerOff(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) BmcStatus(hostList model.HostList) ([]System, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
		runner.RunBmcStatus(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	arr := []System{}
	for m := range ch {
		if m.Status != "success" {
			fmt.Printf("%s\t%s\t%s\n", m.Status, m.Host, m.Msg)
			fmt.Printf("Error querying host: %s\n", m.Host)
			continue
		}
		d := System{}
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

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
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
			fmt.Printf("%s\t%s\t%s\n", m.Status, m.Host, m.Msg)
			fmt.Printf("Error querying host: %s\n", m.Host)
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

func (j *Job) UpdateFirmware(hostList model.HostList, firmwarePaths []string) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
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

func (j *Job) GetJobs(hostList model.HostList) ([]BMCJob, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
		runner.RunGetJobs(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	arr := []BMCJob{}
	for m := range ch {
		if m.Status != "success" {
			fmt.Printf("%s\t%s\t%s\n", m.Status, m.Host, m.Msg)
			continue
		}
		d := BMCJob{}
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

	return arr, nil
}

func (j *Job) ClearJobs(hostList model.HostList) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
		runner.RunClearJobs(host, ch)

		if (i+1)%j.fanout == 0 {
			time.Sleep(j.delay)
			continue
		}
	}

	runner.Wait()
	close(ch)

	return FormatOutput(ch)

}

func (j *Job) PowerCycleBmc(hostList model.HostList) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
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

func (j *Job) ClearSel(hostList model.HostList) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
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

func (j *Job) BmcAutoConfigure(hostList model.HostList) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
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

func (j *Job) BmcImportConfiguration(hostList model.HostList, shutdownType, file string) ([]JobMessage, error) {
	runner := newJobRunner(j)

	ch := make(chan JobMessage, len(hostList))
	for i, host := range hostList {
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
