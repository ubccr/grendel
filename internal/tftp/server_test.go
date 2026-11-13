// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tftp

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/pin/tftp/v3"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/firmware"
	"github.com/ubccr/grendel/pkg/model"
)

// a build in the firmware directory is what a node downloads, and its token names the build the same way the dhcp server does
func TestServesTheFirmwareDirectory(t *testing.T) {
	old := viper.Get("provision.secret")
	viper.Set("provision.secret", "0123456789abcdef0123456789abcdef")
	t.Cleanup(func() { viper.Set("provision.secret", old) })

	dir := t.TempDir()
	want := testPE(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "snponly-x86_64.efi"), want, 0o644))

	s, err := NewServer(nil, "", dir)
	require.NoError(t, err)

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	go s.srv.Serve(conn)
	t.Cleanup(s.srv.Shutdown)

	token, err := model.NewFirmwareToken("aa:bb:cc:dd:ee:ff", firmware.SNPONLYx86_64)
	require.NoError(t, err)

	c, err := tftp.NewClient(conn.LocalAddr().String())
	require.NoError(t, err)
	wt, err := c.Receive(token, "octet")
	require.NoError(t, err)
	var got bytes.Buffer
	_, err = wt.WriteTo(&got)
	require.NoError(t, err)
	assert.Equal(t, want, got.Bytes())
}

func TestNewServerRefusesABadFirmwareDirectory(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "snponly-x86_64.efi"), []byte("not a pe"), 0o644))

	_, err := NewServer(nil, "", dir)
	assert.ErrorContains(t, err, "not an EFI binary", "grendeld fails to start rather than serving a file no node can boot")
}

// testPE is the smallest PE32+ image debug/pe accepts, enough for the firmware directory to take it as an EFI build
func testPE(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	dos := make([]byte, 64)
	copy(dos, "MZ")
	binary.LittleEndian.PutUint32(dos[0x3c:], 64)
	buf.Write(dos)
	buf.WriteString("PE\x00\x00")
	opt := pe.OptionalHeader64{Magic: 0x20b, NumberOfRvaAndSizes: 16}
	require.NoError(t, binary.Write(&buf, binary.LittleEndian, pe.FileHeader{Machine: pe.IMAGE_FILE_MACHINE_AMD64, SizeOfOptionalHeader: uint16(binary.Size(opt))}))
	require.NoError(t, binary.Write(&buf, binary.LittleEndian, opt))
	return buf.Bytes()
}
