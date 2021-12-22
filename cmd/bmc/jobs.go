// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package bmc

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/korovkin/limiter"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/bmc"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/model"
)

type JobRunner struct {
	limit *limiter.ConcurrencyLimiter
}

func NewJobRunner(fanout int) *JobRunner {
	return &JobRunner{limit: limiter.NewConcurrencyLimiter(fanout)}
}

func (j *JobRunner) Wait() {
	j.limit.Wait()
}

func (j *JobRunner) RunStatus(host *model.Host) {
	j.limit.Execute(func() {
		sysmgr, err := systemMgr(host)
		if err != nil {
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host.Name,
				"ID":   host.ID,
			}).Error("Failed to connect to BMC")
			return
		}
		defer sysmgr.Logout()

		system, err := sysmgr.GetSystem()
		if err != nil {
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host.Name,
				"ID":   host.ID,
			}).Error("Failed to fetch system info from BMC")
			return
		}

		if system.Name == "" {
			system.Name = host.Name
		}

		if !statusLong {
			fmt.Printf("%s\t%s\t%s\n",
				host.Name,
				system.PowerStatus,
				system.BIOSVersion)

			return
		}

		rec := make(map[string]*bmc.System, 1)
		rec[host.Name] = system

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")

		if err := enc.Encode(rec); err != nil {
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host.Name,
				"ID":   host.ID,
			}).Error("Failed to encode json")
		}
	})
}

func (j *JobRunner) RunNetBoot(host *model.Host, reboot bool) {
	j.limit.Execute(func() {
		sysmgr, err := systemMgr(host)
		if err != nil {
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host.Name,
				"ID":   host.ID,
			}).Error("Failed to connect to BMC")
			return
		}
		defer sysmgr.Logout()

		err = sysmgr.EnablePXE()
		if err != nil {
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host.Name,
				"ID":   host.ID,
			}).Error("Failed to enabel PXE on next boot")
			return
		}

		if reboot {
			err = sysmgr.PowerCycle()
			if err != nil {
				cmd.Log.WithFields(logrus.Fields{
					"err":  err,
					"name": host.Name,
					"ID":   host.ID,
				}).Error("Failed to power cycle node")
				return
			}
		}

		fmt.Printf("%s: OK\n", host.Name)
	})
}

func (j *JobRunner) RunPower(host *model.Host, powerType int) {
	j.limit.Execute(func() {
		sysmgr, err := systemMgr(host)
		if err != nil {
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host.Name,
				"ID":   host.ID,
			}).Error("Failed to connect to BMC")
			return
		}
		defer sysmgr.Logout()

		switch powerType {
		case PowerCycle:
			err = sysmgr.PowerCycle()
		case PowerOn:
			err = sysmgr.PowerOn()
		case PowerOff:
			err = sysmgr.PowerOff()
		default:
			err = fmt.Errorf("Invalid power type provided: %d", powerType)
		}

		if err != nil {
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host.Name,
				"ID":   host.ID,
			}).Error("Failed to power cycle node")
			return
		}

		fmt.Printf("%s: OK\n", host.Name)
	})
}
