package model

import (
	"fmt"
	"net"
)

// Architecture describes a kind of CPU architecture.
type Architecture int

// Architecture types that Pixiecore knows how to boot.
//
// These architectures are self-reported by the booting machine. The
// machine may support additional execution modes. For example, legacy
// PC BIOS reports itself as an ArchIA32, but may also support ArchX64
// execution.
const (
	// ArchIA32 is a 32-bit x86 machine. It _may_ also support X64
	// execution, but Pixiecore has no way of knowing.
	ArchIA32 Architecture = iota
	// ArchX64 is a 64-bit x86 machine (aka amd64 aka X64).
	ArchX64
)

func (a Architecture) String() string {
	switch a {
	case ArchIA32:
		return "IA32"
	case ArchX64:
		return "X64"
	default:
		return "Unknown architecture"
	}
}

// A Machine describes a machine that is attempting to boot.
type Machine struct {
	MAC  net.HardwareAddr
	Arch Architecture
}

func NewMachineFromDHCP(macaddr net.HardwareAddr, fwt uint16) (*Machine, error) {
	mach := &Machine{MAC: macaddr}

	// Basic architecture identification, based purely on the PXE architecture
	// option.
	switch fwt {
	case 0:
		mach.Arch = ArchIA32
	case 6:
		mach.Arch = ArchIA32
	case 7:
		mach.Arch = ArchX64
	case 9:
		mach.Arch = ArchX64
	default:
		return nil, fmt.Errorf("unsupported client architecture type: %d", fwt)
	}

	return mach, nil
}
