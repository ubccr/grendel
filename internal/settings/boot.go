// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

var (
	TFTPListen = register[string](Option{
		Key:      "tftp.listen",
		Section:  "boot",
		Summary:  "TFTP listen address",
		Help:     "Address and UDP port the TFTP server binds to. It hands booting hosts their iPXE loader.",
		Kind:     KindString,
		Validate: validAddr,
		Default:  "0.0.0.0:69",
		Required: true,
		Apply:    ApplyRestart,
		Flag:     "tftp-listen",
	})
	TFTPFirmwareDir = register[string](Option{
		Key:      "tftp.firmware_dir",
		Section:  "boot",
		Summary:  "iPXE firmware directory",
		Help:     "Directory of iPXE EFI builds to serve in place of the embedded ones, which Secure Boot needs. It is a path on the server, so it is not stored in the database.",
		Kind:     KindPath,
		Default:  "",
		External: true,
		Flag:     "tftp-firmware-dir",
	})
	PXEListen = register[string](Option{
		Key:      "pxe.listen",
		Section:  "boot",
		Summary:  "PXE listen address",
		Help:     "Address and UDP port the PXE server binds to. It answers the boot requests of hosts whose address comes from another DHCP server.",
		Kind:     KindString,
		Validate: validAddr,
		Default:  "0.0.0.0:4011",
		Required: true,
		Apply:    ApplyRestart,
		Flag:     "pxe-listen",
	})
)
