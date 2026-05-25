//go:build pxe

package firmware

import (
	_ "embed"
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
