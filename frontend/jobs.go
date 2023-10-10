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

package frontend

import (
	"fmt"

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

func (j *JobRunner) RunConfigure(host *model.Host, ch chan string) {
	j.limit.Execute(func() {
		ip := host.InterfaceBMC().AddrString()
		err := bmc.ConfigureIdrac(ip)
		msg := "Success"

		if err != nil {
			msg = fmt.Sprintf("Error - %s", err)
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host,
			}).Error("Failed to connect to BMC")
		}
		ch <- fmt.Sprintf("%s: %s", host.Name, msg)
	})
}
func (j *JobRunner) RunReboot(host *model.Host, ch chan string) {
	j.limit.Execute(func() {
		ip := host.InterfaceBMC().AddrString()
		err := bmc.RebootHost(ip)
		msg := "Success"

		if err != nil {
			msg = fmt.Sprintf("Error - %s", err)
			cmd.Log.WithFields(logrus.Fields{
				"err":  err,
				"name": host,
			}).Error("Failed to connect to BMC")
		}
		ch <- fmt.Sprintf("%s: %s", host.Name, msg)
	})
}
