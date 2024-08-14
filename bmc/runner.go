package bmc

import (
	"encoding/json"
	"fmt"

	"github.com/korovkin/limiter"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/model"
)

type jobRunner struct {
	limit    *limiter.ConcurrencyLimiter
	user     string
	pass     string
	insecure bool
}

type JobMessage struct {
	Status       string
	Host         string
	Msg          string
	RedfishError RedfishError
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

func (r *jobRunner) RunPowerCycle(host *model.Host, ch chan JobMessage, bootOverride string) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		ip := host.InterfaceBMC().AddrString()
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.PowerCycle(bootOverride)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = "Sent power cycle command"
	})
}

func (r *jobRunner) RunPowerOn(host *model.Host, ch chan JobMessage, bootOverride string) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		ip := host.InterfaceBMC().AddrString()
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.PowerOn(bootOverride)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = "Sent power on command"
	})
}

func (r *jobRunner) RunPowerOff(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		ip := host.InterfaceBMC().AddrString()
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.PowerOff()
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = "Sent power off command"
	})
}

func (r *jobRunner) RunBmcStatus(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		data := &System{}
		ip := host.InterfaceBMC().AddrString()
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

func (r *jobRunner) RunPowerCycleBmc(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		ip := host.InterfaceBMC().AddrString()
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

func (r *jobRunner) RunClearSel(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		ip := host.InterfaceBMC().AddrString()
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

func (r *jobRunner) RunBmcAutoConfigure(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		ip := host.InterfaceBMC().AddrString()
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

func (r *jobRunner) RunBmcImportConfiguration(host *model.Host, ch chan JobMessage, shutdownType, file string) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		ip := host.InterfaceBMC().AddrString()
		token, err := model.NewBootToken(host.ID.String(), host.InterfaceBMC().MAC.String())
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
