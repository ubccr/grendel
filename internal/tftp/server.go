// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tftp

import (
	"context"
	"time"

	"github.com/pin/tftp/v3"
	"github.com/ubccr/grendel/internal/firmware"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/store"
)

var log = logger.GetLogger("TFTP")

type Server struct {
	Address string
	DB      store.Store
	// firmware is the iPXE builds served, read once at start so a node never gets a half written file
	firmware firmware.Set
	srv      *tftp.Server
}

// NewServer serves the embedded iPXE builds, with any found in firmwareDir in their place
func NewServer(db store.Store, address, firmwareDir string) (*Server, error) {
	fw, err := firmware.Load(firmwareDir)
	if err != nil {
		return nil, err
	}
	logFirmware(fw, firmwareDir)

	s := &Server{DB: db, Address: address, firmware: fw}

	s.srv = tftp.NewServer(s.ReadHandler, nil)
	s.srv.SetTimeout(2 * time.Second)

	return s, nil
}

func (s *Server) Serve() error {
	log.Infof("Server listening on: %s", s.Address)
	return s.srv.ListenAndServe(s.Address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.srv.Shutdown()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}
}

// logFirmware says where each build comes from, warning about what would make a node fail to boot: no firmware at all, an EFI build in the directory that is unsigned or older than this grendeld's, and a BIOS build there that is ignored
func logFirmware(fw firmware.Set, dir string) {
	if len(fw) == 0 {
		log.Warn("No iPXE firmware to serve, this grendeld was built without the pxe tag and no firmware directory has any")
		return
	}
	if dir == "" {
		return
	}
	for _, b := range firmware.Builds() {
		f, ok := fw[b]
		switch {
		case !ok || !b.IsEFI():
			continue
		case f.Path == "":
			log.Infof("Serving the embedded %s, %s has none", b, dir)
		case !f.Signed:
			log.Warnf("Serving %s unsigned, secure boot will refuse it", f.Path)
		case f.Version == firmware.VersionOutdated:
			log.Warnf("Serving %s, which is not this grendeld's %s, export and sign it again after an upgrade", f.Path, b)
		default:
			log.Infof("Serving %s", f.Path)
		}
	}
	for _, path := range firmware.Ignored(dir) {
		log.Warnf("Ignoring %s, BIOS builds are always served from the embedded copy", path)
	}
}
