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

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		}
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

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		}
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

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		}
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
		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
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

func (r *jobRunner) RunGetFirmware(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		data := Firmware{}
		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		data.CurrentFirmwares, err = r.GetFirmware()
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		if r.service.Vendor == "Dell" {
			sys, err := r.GetSystem()
			if err != nil {
				m.Msg = fmt.Sprintf("%s", err)
				return
			}
			data.SystemID = fmt.Sprintf("%04X", sys.OEM.Dell.DellSystem.SystemID)
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

func (jr *jobRunner) RunUpdateFirmware(host *model.Host, ch chan JobMessage, firmwarePaths []string) {
	jr.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		data := FirmwareUpdate{}
		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		}
		r, err := NewRedfishClient(ip, jr.user, jr.pass, jr.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		jids := make(map[string]string)
		for _, firmwarePath := range firmwarePaths {
			jid, err := r.UpdateFirmware(firmwarePath)
			if err != nil {
				m.Msg = fmt.Sprintf("%s", err)
				return
			}
			jids[firmwarePath] = jid
		}

		// time.Sleep(time.Second * 5)

		// completeJobs := make(map[string]*redfish.Job, 0)
		// scheduledJobs := make(map[string]string, 0)

		// for len(jids) > 0 {
		// 	for firmwarePath, jid := range jids {
		// 		job, err := r.GetJobInfo(jid)
		// 		if err != nil {
		// 			m.Msg = fmt.Sprintf("%s", err)
		// 			return
		// 		}
		// 		if job.JobState == "Completed" {
		// 			completeJobs[firmwarePath] = job
		// 			delete(jids, firmwarePath)
		// 		}
		// 		if job.JobState == "Starting" {
		// 			scheduledJobs[firmwarePath] = jid
		// 			delete(jids, firmwarePath)
		// 		}
		// 	}
		// }
		// if len(scheduledJobs) > 0 {
		// 	err := r.PowerCycle("")
		// 	if err != nil {
		// 		m.Msg = fmt.Sprintf("%s", err)
		// 		return // will returning now cause more issues than it's worth??
		// 	}
		// }

		// for len(scheduledJobs) > 0 {
		// 	for firmwarePath, jid := range scheduledJobs {
		// 		job, err := r.GetJobInfo(jid)
		// 		if err != nil {
		// 			if strings.Contains(err.Error(), "401") {
		// 				r.client.Logout()
		// 				r, err = NewRedfishClient(ip, jr.user, jr.pass, jr.insecure)
		// 				if err != nil {
		// 					m.Msg = fmt.Sprintf("%s", err)
		// 					return
		// 				}
		// 			} else {
		// 				m.Msg = fmt.Sprintf("%s", err)
		// 				return
		// 			}
		// 			time.Sleep(time.Minute)
		// 			break
		// 		}

		// 		if job.JobState == "Completed" {
		// 			completeJobs[firmwarePath] = job
		// 			delete(scheduledJobs, firmwarePath)
		// 		}
		// 	}
		// }

		// data.Jobs = completeJobs

		// for i := 0; i < 30; i++ {
		// 	data.CurrentFirmwares, err = r.GetFirmware()
		// 	if err == nil {
		// 		break
		// 	}
		// 	if strings.Contains(err.Error(), "401") {
		// 		r.client.Logout()
		// 		r, err = NewRedfishClient(ip, jr.user, jr.pass, jr.insecure)
		// 		if err != nil {
		// 			m.Msg = fmt.Sprintf("%s", err)
		// 			return
		// 		}
		// 	}
		// 	time.Sleep(time.Minute)
		// }

		// if r.service.Vendor == "Dell" {
		// 	sys, err := r.GetSystem()
		// 	if err != nil {
		// 		m.Msg = fmt.Sprintf("%s", err)
		// 		return
		// 	}
		// 	data.SystemID = fmt.Sprintf("%04X", sys.OEM.Dell.DellSystem.SystemID)
		// }

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

func (r *jobRunner) RunGetJobs(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		data := BMCJob{}
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

func (r *jobRunner) RunClearJobs(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
		}
		r, err := NewRedfishClient(ip, r.user, r.pass, r.insecure)
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		defer r.client.Logout()

		err = r.ClearJobs()
		if err != nil {
			m.Msg = fmt.Sprintf("%s", err)
			return
		}

		m.Status = "success"
		m.Msg = "Cleared Jobs from the BMC"
	})
}

func (r *jobRunner) RunPowerCycleBmc(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
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

func (r *jobRunner) RunClearSel(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
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

func (r *jobRunner) RunBmcAutoConfigure(host *model.Host, ch chan JobMessage) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		ip := ""
		if bmc != nil {
			ip = bmc.AddrString()
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

func (r *jobRunner) RunBmcImportConfiguration(host *model.Host, ch chan JobMessage, shutdownType, file string) {
	r.limit.Execute(func() {
		m := JobMessage{Status: "error", Host: host.Name}
		defer func() { ch <- m }()

		bmc := host.InterfaceBMC()
		mac := ""
		ip := ""
		if bmc != nil {
			mac = bmc.MAC.String()
			ip = bmc.AddrString()
		}
		token, err := model.NewBootToken(host.ID.String(), mac)
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
