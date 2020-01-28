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
