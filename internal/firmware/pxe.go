//go:build !pxe

package firmware

var (
	ipxeBin      []byte
	efi386Bin    []byte
	efi64Bin     []byte
	snpBinX86_64 []byte
	snpBinArm64  []byte
	undiBin      []byte
)
