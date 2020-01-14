package bmc

import (
	"errors"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type Redfish struct {
	config gofish.ClientConfig
	client *gofish.APIClient
}

var resetTypeOrder = []string{
	"PowerCycle",
	"GracefulRestart",
	"ForceRestart",
}

func NewRedfish(endpoint, user, pass string, insecure bool) (*Redfish, error) {
	config := gofish.ClientConfig{
		Endpoint: endpoint,
		Username: user,
		Password: pass,
		Insecure: insecure,
	}

	fish, err := gofish.Connect(config)
	if err != nil {
		return nil, err
	}

	return &Redfish{config: config, client: fish}, nil
}

func (r *Redfish) Logout() {
	r.client.Logout()
}

func (r *Redfish) PowerCycle() error {
	service := r.client.Service
	ss, err := service.Systems()
	if err != nil {
		return err
	}

	// XXX Only reset the first supported system?
	for _, system := range ss {
		for _, resetType := range resetTypeOrder {
			for _, rt := range system.SupportedResetTypes {
				if resetType == string(rt) {
					err = system.Reset(rt)
					if err != nil {
						return err
					}
					return nil
				}
			}
		}
	}

	return errors.New("Failed to find a supported reset type")
}

func (r *Redfish) EnablePXE() error {
	service := r.client.Service
	ss, err := service.Systems()
	if err != nil {
		return err
	}

	bootOverride := redfish.Boot{
		BootSourceOverrideTarget:  redfish.PxeBootSourceOverrideTarget,
		BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
	}

	for _, system := range ss {
		err := system.SetBoot(bootOverride)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Redfish) GetSystem() (*System, error) {
	service := r.client.Service
	ss, err := service.Systems()
	if err != nil {
		return nil, err
	}

	if len(ss) == 0 {
		return nil, errors.New("Failed to find system")
	}

	sys := ss[0]

	system := &System{
		Name:           sys.HostName,
		BIOSVersion:    sys.BIOSVersion,
		SerialNumber:   sys.SKU,
		Manufacturer:   sys.Manufacturer,
		PowerStatus:    string(sys.PowerState),
		Health:         string(sys.Status.Health),
		TotalMemory:    sys.MemorySummary.TotalSystemMemoryGiB,
		ProcessorCount: sys.ProcessorSummary.LogicalProcessorCount,
		BootNext:       sys.Boot.BootNext,
		BootOrder:      sys.Boot.BootOrder,
	}

	return system, nil
}
