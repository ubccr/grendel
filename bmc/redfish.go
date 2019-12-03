package bmc

import (
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type Redfish struct {
	config gofish.ClientConfig
	client *gofish.APIClient
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

	for _, system := range ss {
		err = system.Reset("GracefulRestart")
		if err != nil {
			return err
		}
	}

	return nil
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
