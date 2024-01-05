package bmc

import (
	"errors"
	"net"
	"strings"

	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/ubccr/grendel/util"
)

// PowerCycle2 will power cycle the BMC
func (r *Redfish2) PowerCycle2(bootOverride redfish.Boot) error {
	ss, err := r.service.Systems()
	if err != nil {
		return err
	}

	defer r.client.Logout()

	// TODO: support GracefulRestartResetType?

	err = r.bootOverride(bootOverride)
	if err != nil {
		return err
	}

	for _, s := range ss {
		err := *new(error)

		switch s.PowerState {
		case redfish.OnPowerState:
			err = s.Reset(redfish.ForceRestartResetType)
		case redfish.OffPowerState:
			err = s.Reset(redfish.OnResetType)
		// Can you ForceRestart from a paused state?
		case redfish.PausedPowerState:
			err = s.Reset(redfish.ForceRestartResetType)
		case redfish.PoweringOnPowerState:
			err = s.Reset(redfish.ForceRestartResetType)
		case redfish.PoweringOffPowerState:
			err = s.Reset(redfish.ForceRestartResetType)
		default:
			// TODO: Log powerstate
			err = errors.New("failed to match current power state")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// PowerOn2 will power on the BMC
func (r *Redfish2) PowerOn2(bootOverride redfish.Boot) error {
	ss, err := r.service.Systems()
	if err != nil {
		return err
	}

	defer r.client.Logout()

	// TODO: support GracefulRestartResetType?

	err = r.bootOverride(bootOverride)
	if err != nil {
		return err
	}

	for _, s := range ss {
		err := *new(error)

		switch s.PowerState {
		case redfish.OnPowerState:
			err = nil
		case redfish.OffPowerState:
			err = s.Reset(redfish.OnResetType)
		case redfish.PausedPowerState:
			err = s.Reset(redfish.ResumeResetType)
		case redfish.PoweringOnPowerState:
			err = nil
		case redfish.PoweringOffPowerState:
			err = s.Reset(redfish.ForceOnResetType)
		default:
			// TODO: Log powerstate
			err = errors.New("failed to match current power state")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// PowerOff2 will power off the BMC
func (r *Redfish2) PowerOff2() error {
	ss, err := r.service.Systems()
	if err != nil {
		return err
	}

	defer r.client.Logout()

	for _, s := range ss {
		err := *new(error)

		switch s.PowerState {
		case redfish.OnPowerState:
			err = s.Reset(redfish.ForceOffResetType)
		case redfish.OffPowerState:
			err = nil
		case redfish.PausedPowerState:
			err = s.Reset(redfish.ForceOffResetType)
		case redfish.PoweringOnPowerState:
			err = s.Reset(redfish.ForceOffResetType)
		case redfish.PoweringOffPowerState:
			err = nil
		default:
			// TODO: Log powerstate
			err = errors.New("failed to match current power state")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// bootOverride will set the boot override target
func (r *Redfish2) bootOverride(bootOverride redfish.Boot) error {
	ss, err := r.service.Systems()
	if err != nil {
		return err
	}

	for _, s := range ss {
		err := s.SetBoot(bootOverride)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Redfish2) GetSystem2() (*System2, error) {
	ss, err := r.service.Systems()
	if err != nil {
		return nil, err
	}

	if len(ss) == 0 {
		return nil, errors.New("failed to find system")
	}

	sys := ss[0]

	system := &System2{
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

func (r *Redfish2) PowerCycleBmc() error {
	ms, err := r.service.Managers()
	if err != nil {
		return err
	}

	for _, m := range ms {
		err := m.Reset(redfish.GracefulRestartResetType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Redfish2) ClearSel() error {
	ms, err := r.service.Managers()
	if err != nil {
		return err
	}

	for _, m := range ms {
		ls, err := m.LogServices()
		if err != nil {
			return err
		}
		for _, l := range ls {
			err := l.ClearLog()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Redfish2) BmcAutoConfigure() error {
	// TODO: Submit PR to gofish to support this natively?

	type attributes struct {
		NIC1AutoConfig string `json:"NIC.1.AutoConfig"`
	}
	type payload struct {
		Attributes attributes
	}

	p := payload{
		Attributes: attributes{
			NIC1AutoConfig: "Enable Once",
		},
	}

	return r.service.Patch("/redfish/v1/Managers/iDRAC.Embedded.1/Attributes", p)
}

func (r *Redfish2) BmcImportConfiguration(shutdownType, path, file string) error {
	// TODO: Submit PR to gofish to support this natively?

	type shareParameters struct {
		Target                   []string
		ShareType                string
		IPAddress                string
		FileName                 string
		ShareName                string
		PortNumber               string
		IgnoreCertificateWarning string
	}
	type payload struct {
		HostPowerState  string
		ShutdownType    string
		ImportBuffer    string
		ShareParameters shareParameters
	}

	rawScheme := viper.GetString("provision.scheme")
	scheme := strings.ToUpper(rawScheme)

	ip, err := util.GetFirstExternalIPFromInterfaces()
	if err != nil {
		return err
	}
	_, port, err := net.SplitHostPort(viper.GetString("provision.listen"))
	if err != nil {
		return err
	}
	viper.SetDefault("bmc.config_ignore_certificate_warning", "Disabled")
	icw := viper.GetString("bmc.config_ignore_certificate_warning")

	p := payload{
		HostPowerState: "On",
		ShutdownType:   shutdownType,
		ShareParameters: shareParameters{
			Target:                   []string{"ALL"},
			ShareType:                scheme,
			IPAddress:                ip.String(),
			PortNumber:               port,
			FileName:                 file,
			ShareName:                path,
			IgnoreCertificateWarning: icw,
		},
	}

	return r.service.Post("/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfiguration", p)
}
