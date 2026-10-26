// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package firmware

import (
	"testing"

	"github.com/insomniacslk/dhcp/iana"
	"github.com/stretchr/testify/assert"
)

func TestDetectArch(t *testing.T) {
	assert := assert.New(t)

	tests := map[iana.Arch]string{
		iana.INTEL_X86PC: "x86_64",
		iana.EFI_BC:      "x86_64",
		iana.EFI_X86_64:  "x86_64",
		iana.EFI_IA32:    "i386",
		iana.EFI_ARM64:   "aarch64",
	}

	for arch, want := range tests {
		got, err := DetectArch(iana.Archs{arch})
		if assert.NoErrorf(err, "unexpected error for %s", arch) {
			assert.Equal(want, got)
		}
	}

	_, err := DetectArch(iana.Archs{iana.EFI_ARM32})
	assert.Error(err)

	_, err = DetectArch(iana.Archs{})
	assert.Error(err)
}
