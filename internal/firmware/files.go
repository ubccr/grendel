// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package firmware

import (
	"bytes"
	"crypto/sha256"
	"debug/pe"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// File is one build as Grendel serves it. Path is where it was read from, empty for the copy embedded in the binary
type File struct {
	Data   []byte
	Path   string
	Signed bool
	// Version says whether a file from the directory is the embedded build, signatures aside. A copy signed before a Grendel upgrade is VersionOutdated
	Version Version
}

// Version compares a file from the firmware directory with the build embedded in grendeld
type Version int

const (
	// VersionUnknown is an embedded build, or a file with no embedded build to compare it with
	VersionUnknown Version = iota
	VersionCurrent
	VersionOutdated
)

// Set is the build served for each firmware type
type Set map[Build]File

// IsEFI reports whether a build is a UEFI binary, the only kind secure boot checks and so the only kind the firmware directory holds
func (b Build) IsEFI() bool {
	switch b {
	case EFI386, EFI64, SNPONLYx86_64, SNPONLYarm64:
		return true
	}
	return false
}

// Builds lists every build in a fixed order, so listings and exports read the same every time
func Builds() []Build {
	builds := make([]Build, 0, len(BuildToStringMap))
	for b := range BuildToStringMap {
		builds = append(builds, b)
	}
	slices.Sort(builds)
	return builds
}

// Embedded is the builds compiled into this binary. Without the pxe build tag there are none
func Embedded() Set {
	set := Set{}
	for _, b := range Builds() {
		data := b.ToBytes()
		if len(data) == 0 {
			continue
		}
		f := File{Data: data}
		if b.IsEFI() {
			f.Signed, _ = Signed(data)
		}
		set[b] = f
	}
	return set
}

// Load starts from the embedded builds and replaces each EFI build found in dir under its build name, such as snponly-x86_64.efi. A build missing from dir keeps the embedded copy, so dir only needs the builds that were signed. BIOS builds are always the embedded ones, see Ignored. An empty dir is the embedded set. A file that is there but is not an EFI binary is an error rather than something to find out about when a node fails to boot
func Load(dir string) (Set, error) {
	set := Embedded()
	if dir == "" {
		return set, nil
	}

	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("firmware directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("firmware directory %s is not a directory", dir)
	}

	for _, b := range Builds() {
		if !b.IsEFI() {
			continue
		}
		path := filepath.Join(dir, b.String())
		data, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}

		f := File{Data: data, Path: path}
		if f.Signed, err = Signed(data); err != nil {
			return nil, fmt.Errorf("%s is not an EFI binary: %w", path, err)
		}
		if embedded, ok := set[b]; ok {
			f.Version = VersionCurrent
			if imageDigest(data) != imageDigest(embedded.Data) {
				f.Version = VersionOutdated
			}
		}
		set[b] = f
	}
	return set, nil
}

// Ignored lists the BIOS builds found in dir, which Grendel does not serve from there since secure boot does not apply to them
func Ignored(dir string) []string {
	if dir == "" {
		return nil
	}
	var out []string
	for _, b := range Builds() {
		if b.IsEFI() {
			continue
		}
		if path := filepath.Join(dir, b.String()); fileExists(path) {
			out = append(out, path)
		}
	}
	return out
}

// Bytes is the firmware served for a build, nil when there is none
func (s Set) Bytes(b Build) []byte {
	return s[b].Data
}

// security is the PE's certificate table entry, and where it sits in the file. The table holds the Authenticode signatures, appended after the image
type security struct {
	dir    pe.DataDirectory
	offset int // of the directory entry itself
	sum    int // of the optional header checksum
}

func readSecurity(data []byte) (security, error) {
	f, err := pe.NewFile(bytes.NewReader(data))
	if err != nil {
		return security{}, err
	}
	defer f.Close()

	if len(data) < 0x40 {
		return security{}, errors.New("too short for a PE")
	}
	// the optional header follows the PE signature and the 20 byte file header
	opt := int(binary.LittleEndian.Uint32(data[0x3c:])) + 4 + 20

	// NumberOfRvaAndSizes is how many of the directories are in use, and comes from the file, so it is kept within the array
	var dirs []pe.DataDirectory
	var dirsAt int
	switch h := f.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		dirs, dirsAt = h.DataDirectory[:min(h.NumberOfRvaAndSizes, uint32(len(h.DataDirectory)))], opt+96
	case *pe.OptionalHeader64:
		dirs, dirsAt = h.DataDirectory[:min(h.NumberOfRvaAndSizes, uint32(len(h.DataDirectory)))], opt+112
	default:
		return security{}, errors.New("no optional header")
	}

	s := security{sum: opt + 64, offset: dirsAt + 8*pe.IMAGE_DIRECTORY_ENTRY_SECURITY}
	if len(dirs) > pe.IMAGE_DIRECTORY_ENTRY_SECURITY {
		s.dir = dirs[pe.IMAGE_DIRECTORY_ENTRY_SECURITY]
	}
	return s, nil
}

// Signed reports whether a PE binary carries an Authenticode signature, which is what UEFI secure boot checks against db. It only says a signature is present, not whose it is or that it verifies
func Signed(data []byte) (bool, error) {
	s, err := readSecurity(data)
	if err != nil {
		return false, err
	}
	return s.dir.Size > 0, nil
}

// imageDigest hashes a PE with what signing changes left out: the checksum, the certificate table entry, the signatures appended at the end, and the zero padding a signer adds before them. A signed copy and the build it was signed from hash the same, the way Authenticode itself ignores those parts. A file that is not a PE hashes whole
func imageDigest(data []byte) [32]byte {
	s, err := readSecurity(data)
	if err != nil || s.offset+8 > len(data) || s.sum+4 > len(data) {
		return sha256.Sum256(data)
	}

	image := slices.Clone(data)
	clear(image[s.sum : s.sum+4])
	clear(image[s.offset : s.offset+8])
	if start := int(s.dir.VirtualAddress); s.dir.Size > 0 && start <= len(image) {
		image = image[:start]
	}
	return sha256.Sum256(bytes.TrimRight(image, "\x00"))
}

// Export writes the embedded EFI builds to dir, creating it if need be, and returns the paths written. It refuses to overwrite a file unless force is set, since the file there may be one already signed, and checks them all first so an export is never left half done
func Export(dir string, force bool) ([]string, error) {
	embedded := Embedded()
	var builds []Build
	for _, b := range Builds() {
		if _, ok := embedded[b]; ok && b.IsEFI() {
			builds = append(builds, b)
		}
	}
	if len(builds) == 0 {
		return nil, errors.New("this grendeld was built without embedded firmware, build it with the pxe tag")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	paths := make([]string, len(builds))
	for i, b := range builds {
		paths[i] = filepath.Join(dir, b.String())
	}

	if !force {
		var exists []string
		for _, path := range paths {
			if fileExists(path) {
				exists = append(exists, path)
			}
		}
		if len(exists) > 0 {
			return nil, fmt.Errorf("%s already exists, pass --force to overwrite", strings.Join(exists, ", "))
		}
	}

	for i, b := range builds {
		if err := os.WriteFile(paths[i], embedded[b].Data, 0o644); err != nil {
			return paths[:i], err
		}
	}
	return paths, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
