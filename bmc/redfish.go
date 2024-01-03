package bmc

import (
	"errors"

	"github.com/stmcginnis/gofish/redfish"
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

func (r *Redfish2) PowerCycleBmc() error {
	ms, err := r.service.Managers()
	if err != nil {
		return err
	}

	for _, m := range ms {
		err := m.Reset(redfish.PowerCycleResetType)
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
