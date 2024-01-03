package bmc

import (
	"fmt"

	"github.com/korovkin/limiter"
	"github.com/stmcginnis/gofish/redfish"
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

func (j *JobRunner) RunPowerCycle(host *model.Host, ch chan string, bootOverride redfish.Boot) {
	j.limit.Execute(func() {
		status := "success"
		msg := "Sent power cycle command"

		ip := host.InterfaceBMC().AddrString()
		r, err := NewClient(ip)
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
			ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
			return
		}
		err = r.PowerCycle2(bootOverride)
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
		}

		ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
	})
}

func (j *JobRunner) RunPowerOn(host *model.Host, ch chan string, bootOverride redfish.Boot) {
	j.limit.Execute(func() {
		status := "success"
		msg := "Sent power on command"

		ip := host.InterfaceBMC().AddrString()
		r, err := NewClient(ip)
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
			ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
			return
		}
		err = r.PowerOn2(bootOverride)
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
		}

		ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
	})
}

func (j *JobRunner) RunPowerOff(host *model.Host, ch chan string) {
	j.limit.Execute(func() {
		status := "success"
		msg := "Sent power off command"

		ip := host.InterfaceBMC().AddrString()
		r, err := NewClient(ip)
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
			ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
			return
		}
		err = r.PowerOff2()
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
		}

		ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
	})
}

func (j *JobRunner) RunPowerCycleBmc(host *model.Host, ch chan string) {
	j.limit.Execute(func() {
		status := "success"
		msg := "Sent bmc reboot command"

		ip := host.InterfaceBMC().AddrString()
		r, err := NewClient(ip)
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
			ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
			return
		}
		err = r.PowerCycleBmc()
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
		}

		ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
	})
}

func (j *JobRunner) RunClearSel(host *model.Host, ch chan string) {
	j.limit.Execute(func() {
		status := "success"
		msg := "Sent clear sel command"

		ip := host.InterfaceBMC().AddrString()
		r, err := NewClient(ip)
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
			ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
			return
		}
		err = r.ClearSel()
		if err != nil {
			status = "error"
			msg = fmt.Sprintf("%s", err)
		}

		ch <- fmt.Sprintf("%s|%s|%s", status, host.Name, msg)
	})
}
