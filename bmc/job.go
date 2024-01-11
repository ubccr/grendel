package bmc

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/model"
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
			rErr, err := ParseRedfishError(errors.New(m.Msg))
			if err != nil {
				return nil, err
			}
			m.RedfishError = rErr
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

	arr := []JobMessage{}
	for m := range ch {
		arr = append(arr, m)
	}

	return arr, nil
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

	arr := []JobMessage{}
	for m := range ch {
		arr = append(arr, m)
	}

	return arr, nil
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

	arr := []JobMessage{}
	for m := range ch {
		arr = append(arr, m)
	}

	return arr, nil
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

	arr := []JobMessage{}
	for m := range ch {
		arr = append(arr, m)
	}

	return arr, nil
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

	arr := []JobMessage{}
	for m := range ch {
		arr = append(arr, m)
	}

	return arr, nil
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

	arr := []JobMessage{}
	for m := range ch {
		arr = append(arr, m)
	}

	return arr, nil
}
