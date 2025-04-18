// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package firmware

import (
	_ "embed"
	"fmt"

	"github.com/insomniacslk/dhcp/iana"
)

type Build int

const (
	IPXE Build = iota + 1
	EFI386
	EFI64
	SNPONLYx86_64
	SNPONLYarm64
	UNDI
	GRENDEL
)

//go:embed bin/ipxe.pxe
var ipxeBin []byte

//go:embed bin/ipxe-i386.efi
var efi386Bin []byte

//go:embed bin/ipxe-x86_64.efi
var efi64Bin []byte

//go:embed bin/snponly-x86_64.efi
var snpBinX86_64 []byte

//go:embed bin/snponly-arm64.efi
var snpBinArm64 []byte

//go:embed bin/undionly.kpxe
var undiBin []byte

// buildToStringMap maps a Build to a binary build name
var BuildToStringMap = map[Build]string{
	IPXE:          "ipxe.pxe",
	EFI386:        "ipxe-i386.efi",
	EFI64:         "ipxe-x86_64.efi",
	SNPONLYx86_64: "snponly-x86_64.efi",
	SNPONLYarm64:  "snponly-arm64.efi",
	UNDI:          "undionly.kpxe",
}

func NewFromString(b string) Build {
	for k, v := range BuildToStringMap {
		if v == b {
			return k
		}
	}

	return Build(0)
}

// String returns a name for a given build.
func (b Build) String() string {
	if bt, ok := BuildToStringMap[b]; ok {
		return bt
	}
	return ""
}

func (b Build) IsNil() bool {
	return int(b) == 0
}

func (b Build) ToBytes() []byte {
	switch b {
	case IPXE:
		return ipxeBin
	case EFI386:
		return efi386Bin
	case EFI64:
		return efi64Bin
	case SNPONLYx86_64:
		return snpBinX86_64
	case SNPONLYarm64:
		return snpBinArm64
	case UNDI:
		return undiBin
	}

	return nil
}

func DetectBuild(archs iana.Archs, userClass string) (Build, error) {
	var build Build

	if len(archs) == 0 {
		return build, fmt.Errorf("no client system architecture types provided")
	}

	//XXX TODO use first arch? what to do if there's more than one??
	arch := archs[0]

	switch arch {
	case iana.INTEL_X86PC: // BIOS Boot
		build = UNDI
	case iana.EFI_IA32: // unverified
		build = EFI386
	case iana.EFI_BC, iana.EFI_X86_64: // UEFI x86_64 Boot
		build = SNPONLYx86_64
	case iana.EFI_ARM64: // UEFI ARM64
		build = SNPONLYarm64
	default:
		return build, fmt.Errorf("unsupported client system architecture type: %d", arch)
	}

	if userClass != "" {
		if userClass == "iPXE" && arch == iana.INTEL_X86PC {
			build = IPXE
		}
		if userClass == "grendel" {
			build = GRENDEL
		}
	}

	return build, nil
}
