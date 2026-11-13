// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package firmware

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPE builds the smallest PE32+ image debug/pe accepts, with an Authenticode certificate table appended when signed, the way a signer adds one. machine stands in for what makes two builds differ
func testPE(t *testing.T, signed bool, machine ...uint16) []byte {
	t.Helper()
	var buf bytes.Buffer
	dos := make([]byte, 64)
	copy(dos, "MZ")
	binary.LittleEndian.PutUint32(dos[0x3c:], 64)
	buf.Write(dos)
	buf.WriteString("PE\x00\x00")

	opt := pe.OptionalHeader64{Magic: 0x20b, NumberOfRvaAndSizes: 16}
	hdr := pe.FileHeader{Machine: pe.IMAGE_FILE_MACHINE_AMD64, SizeOfOptionalHeader: uint16(binary.Size(opt))}
	if len(machine) > 0 {
		hdr.Machine = machine[0]
	}
	// a signer pads the image with zeros before appending the certificate table, which the digest has to see past
	size := buf.Len() + binary.Size(hdr) + binary.Size(opt)
	if signed {
		opt.CheckSum = 0x1234
		opt.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY] = pe.DataDirectory{VirtualAddress: uint32(size + 8), Size: 16}
	}
	require.NoError(t, binary.Write(&buf, binary.LittleEndian, hdr))
	require.NoError(t, binary.Write(&buf, binary.LittleEndian, opt))
	if signed {
		buf.Write(make([]byte, 8))
		buf.WriteString("certificate-data")
	}
	return buf.Bytes()
}

// withEmbedded stands in for the pxe build tag, restoring the package's embedded builds after the test. It swaps package variables, so tests using it cannot run in parallel
func withEmbedded(t *testing.T, snp, undi []byte) {
	t.Helper()
	oldSnp, oldUndi := snpBinX86_64, undiBin
	snpBinX86_64, undiBin = snp, undi
	t.Cleanup(func() { snpBinX86_64, undiBin = oldSnp, oldUndi })
}

func TestSigned(t *testing.T) {
	signed, err := Signed(testPE(t, true))
	require.NoError(t, err)
	assert.True(t, signed)

	signed, err = Signed(testPE(t, false))
	require.NoError(t, err)
	assert.False(t, signed)

	_, err = Signed([]byte("not a pe"))
	assert.Error(t, err)
}

func TestLoadPrefersTheDirectory(t *testing.T) {
	embedded := testPE(t, false)
	withEmbedded(t, embedded, []byte("undi"))

	dir := t.TempDir()
	signed := testPE(t, true)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "snponly-x86_64.efi"), signed, 0o644))

	set, err := Load(dir)
	require.NoError(t, err)

	snp := set[SNPONLYx86_64]
	assert.Equal(t, signed, snp.Data)
	assert.Equal(t, filepath.Join(dir, "snponly-x86_64.efi"), snp.Path)
	assert.True(t, snp.Signed)

	assert.Equal(t, VersionCurrent, snp.Version, "a signed copy of the embedded build is current")

	undi := set[UNDI]
	assert.Equal(t, []byte("undi"), undi.Data, "a build missing from the directory keeps the embedded copy")
	assert.Empty(t, undi.Path)

	_, ok := set[SNPONLYarm64]
	assert.False(t, ok, "a build neither embedded nor in the directory is absent")
	assert.Nil(t, set.Bytes(SNPONLYarm64))
}

func TestLoadRefusesWhatCannotBeServed(t *testing.T) {
	withEmbedded(t, nil, nil)

	cases := map[string][]byte{
		"snponly-x86_64.efi": []byte("not a pe"),
		"ipxe-x86_64.efi":    {},
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), data, 0o644))
			_, err := Load(dir)
			assert.Error(t, err)
		})
	}

	_, err := Load(filepath.Join(t.TempDir(), "missing"))
	assert.Error(t, err, "a configured directory that is not there is an error, not the embedded set")
}

func TestExport(t *testing.T) {
	withEmbedded(t, testPE(t, false), []byte("undi"))
	dir := filepath.Join(t.TempDir(), "fw")

	paths, err := Export(dir, false)
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dir, "snponly-x86_64.efi")}, paths, "only the EFI builds are written")

	_, err = Export(dir, false)
	assert.ErrorContains(t, err, "already exists", "a file there, possibly signed, is not overwritten")

	_, err = Export(dir, true)
	assert.NoError(t, err)

	set, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "snponly-x86_64.efi"), set[SNPONLYx86_64].Path, "an export loads back as the served build")
}

func TestExportWithoutFirmware(t *testing.T) {
	withEmbedded(t, nil, nil)
	_, err := Export(t.TempDir(), false)
	assert.ErrorContains(t, err, "built without embedded firmware")
}

func TestLoadFlagsAnOutdatedBuild(t *testing.T) {
	withEmbedded(t, testPE(t, false), nil)

	dir := t.TempDir()
	// signed, but from a different build than this grendeld embeds, as after an upgrade
	require.NoError(t, os.WriteFile(filepath.Join(dir, "snponly-x86_64.efi"), testPE(t, true, pe.IMAGE_FILE_MACHINE_ARM64), 0o644))

	set, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, VersionOutdated, set[SNPONLYx86_64].Version)
}

func TestLoadWithNothingToCompare(t *testing.T) {
	withEmbedded(t, nil, nil)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "snponly-x86_64.efi"), testPE(t, true), 0o644))

	set, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, VersionUnknown, set[SNPONLYx86_64].Version, "without an embedded build there is nothing to call it current or outdated against")
}

func TestLoadIgnoresBIOSBuilds(t *testing.T) {
	withEmbedded(t, nil, []byte("undi"))

	dir := t.TempDir()
	path := filepath.Join(dir, "undionly.kpxe")
	require.NoError(t, os.WriteFile(path, []byte("custom"), 0o644))

	set, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, []byte("undi"), set.Bytes(UNDI), "a BIOS build is always the embedded one")
	assert.Equal(t, []string{path}, Ignored(dir))
	assert.Nil(t, Ignored(""))
}

func TestImageDigestIgnoresTheSignature(t *testing.T) {
	assert.Equal(t, imageDigest(testPE(t, false)), imageDigest(testPE(t, true)))
	assert.NotEqual(t, imageDigest(testPE(t, false)), imageDigest(testPE(t, false, pe.IMAGE_FILE_MACHINE_ARM64)))
}
