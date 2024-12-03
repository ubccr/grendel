// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/spf13/viper"
)

var (
	machineRegexp = regexp.MustCompile("^[a-zA-Z0-9_][a-zA-Z0-9_]{0,64}$")
)

type OnieOp int

const (
	OnieInstall OnieOp = iota
	OnieUpdate
)

type Onie struct {
	SerialNumber string
	MAC          net.HardwareAddr
	VendorID     uint32
	Machine      string
	MachineRev   uint32
	Arch         string
	SecurityKey  string
	Operation    OnieOp
}

func (op OnieOp) String() string {
	return [...]string{"os-install", "onie-update"}[op]
}

func NewOnieFromHeaders(header http.Header) (*Onie, error) {
	o := &Onie{
		SerialNumber: header.Get("ONIE-SERIAL-NUMBER"),
		SecurityKey:  header.Get("ONIE-SECURITY-KEY"),
		Machine:      header.Get("ONIE-MACHINE"),
		Arch:         header.Get("ONIE-ARCH"),
	}

	var err error
	o.MAC, err = net.ParseMAC(header.Get("ONIE-ETH-ADDR"))
	if err != nil {
		return nil, errors.New("Onie invalid mac address")
	}

	if id, err := strconv.ParseUint(header.Get("ONIE-VENDOR-ID"), 10, 32); err == nil {
		o.VendorID = uint32(id)
	} else {
		return nil, errors.New("Onie invalid vendor id")
	}

	if rev, err := strconv.ParseUint(header.Get("ONIE-MACHINE-REV"), 10, 32); err == nil {
		o.MachineRev = uint32(rev)
	} else {
		return nil, errors.New("Onie invalid machine rev")
	}

	if o.Arch != "x86_64" {
		return nil, errors.New("Onie invalid arch")
	}

	if !machineRegexp.MatchString(o.Machine) {
		return nil, errors.New("Onie invalid machine")
	}

	onieOp := header.Get("ONIE-OPERATION")
	switch {
	case onieOp == "onie-update":
		o.Operation = OnieUpdate
	case onieOp == "os-install":
		o.Operation = OnieInstall
	default:
		return nil, errors.New("Onie invalid operation")
	}

	return o, nil
}

func (o Onie) UpdaterFilePath() string {
	return filepath.Join(
		viper.GetString("provision.repo_dir"),
		"onie",
		fmt.Sprintf("%s-%s-%s-r%d", "onie-updater", o.Arch, o.Machine, o.MachineRev))
}
