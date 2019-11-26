package firmware

import (
	"fmt"
)

// This code was adopted from Pixiecore
// https://github.com/danderson/netboot/blob/master/pixiecore/pixiecore.go

// BootLoader describes a kind of firmware attempting to boot.
type BootLoader int

// The bootloaders that Pixiecore knows how to handle.
const (
	X86PC         BootLoader = iota // "Classic" x86 BIOS with PXE/UNDI support
	EFI32                           // 32-bit x86 processor running EFI
	EFI64                           // 64-bit x86 processor running EFI
	EFIBC                           // 64-bit x86 processor running EFI
	X86Ipxe                         // "Classic" x86 BIOS running iPXE (no UNDI support)
	PixiecoreIpxe                   // Pixiecore's iPXE, which has replaced the underlying firmware
)

var (
	IPXEBin map[BootLoader][]byte
)

func init() {
	IPXEBin = make(map[BootLoader][]byte, 0)
	IPXEBin[X86PC] = MustAsset("undionly.kpxe")
	IPXEBin[EFI32] = MustAsset("ipxe-i386.efi")
	IPXEBin[EFI64] = MustAsset("snponly-x86_64.efi")
	IPXEBin[EFIBC] = MustAsset("snponly-x86_64.efi")
	IPXEBin[X86Ipxe] = MustAsset("ipxe.pxe")
}

func GetBootLoader(bootLoader BootLoader) ([]byte, error) {
	bs, ok := IPXEBin[bootLoader]
	if !ok {
		return nil, fmt.Errorf("unknown firmware type %d", bootLoader)
	}

	return bs, nil
}

func DetectBootLoader(fwt uint16, userClass string) (BootLoader, error) {
	var fwtype BootLoader

	// Basic firmware identification, based purely on the PXE architecture
	// option.
	switch fwt {
	case 0:
		fwtype = X86PC
	case 6:
		fwtype = EFI32
	case 7:
		fwtype = EFI64
	case 9:
		fwtype = EFIBC
	default:
		return fwtype, fmt.Errorf("unsupported client firmware type: %d", fwt)
	}

	// Now, identify special sub-breeds of client firmware based on
	// the user-class option. Note these only change the "firmware
	// type", not the architecture we're reporting to Booters. We need
	// to identify these as part of making the internal chainloading
	// logic work properly.
	if userClass != "" {
		// If the client has had iPXE burned into its ROM (or is a VM
		// that uses iPXE as the PXE "ROM"), special handling is
		// needed because in this mode the client is using iPXE native
		// drivers and chainloading to a UNDI stack won't work.
		if userClass == "iPXE" && fwtype == X86PC {
			fwtype = X86Ipxe
		}
		// If the client identifies as "pixiecore", we've already
		// chainloaded this client to the full-featured copy of iPXE
		// we supply. We have to distinguish this case so we don't
		// loop on the chainload step.
		if userClass == "pixiecore" {
			fwtype = PixiecoreIpxe
		}
	}

	return fwtype, nil
}
