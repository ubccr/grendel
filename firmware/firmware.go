// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

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
	SNPONLY
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
var snpBin []byte

//go:embed bin/undionly.kpxe
var undiBin []byte

// buildToStringMap maps a Build to a binary build name
var buildToStringMap = map[Build]string{
	IPXE:    "ipxe.pxe",
	EFI386:  "ipxe-i386.efi",
	EFI64:   "ipxe-x86_64.efi",
	SNPONLY: "snponly-x86_64.efi",
	UNDI:    "undionly.kpxe",
}

func NewFromString(b string) Build {
	for k, v := range buildToStringMap {
		if v == b {
			return k
		}
	}

	return Build(0)
}

// String returns a name for a given build.
func (b Build) String() string {
	if bt, ok := buildToStringMap[b]; ok {
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
	case SNPONLY:
		return snpBin
	case UNDI:
		return undiBin
	}

	return nil
}

func DetectBuild(archs iana.Archs, userClass string) (Build, error) {
	var build Build

	if archs == nil || len(archs) == 0 {
		return build, fmt.Errorf("No Client System Architecture Types provided")
	}

	//XXX TODO use first arch? what to do if there's more than one??
	arch := archs[0]

	switch arch {
	case iana.INTEL_X86PC:
		build = UNDI
	case iana.EFI_IA32:
		build = EFI386
	case iana.EFI_BC, iana.EFI_X86_64:
		build = EFI64
	default:
		return build, fmt.Errorf("Unsupported Client System Architecture Type: %d", arch)
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
