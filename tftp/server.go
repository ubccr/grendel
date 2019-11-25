package tftp

import (
	"time"

	"github.com/pin/tftp"
	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
)

type Server struct {
	FirmwareBin map[model.Firmware][]byte
	Address     string
	srv         *tftp.Server
}

func NewServer(address string) (*Server, error) {
	s := &Server{Address: address}

	s.FirmwareBin = make(map[model.Firmware][]byte, 0)
	s.FirmwareBin[model.FirmwareX86PC] = firmware.MustAsset("undionly.kpxe")
	s.FirmwareBin[model.FirmwareEFI32] = firmware.MustAsset("ipxe-i386.efi")
	s.FirmwareBin[model.FirmwareEFI64] = firmware.MustAsset("snponly-x86_64.efi")
	s.FirmwareBin[model.FirmwareEFIBC] = firmware.MustAsset("snponly-x86_64.efi")
	s.FirmwareBin[model.FirmwareX86Ipxe] = firmware.MustAsset("ipxe.pxe")

	s.srv = tftp.NewServer(s.ReadHandler, nil)
	s.srv.SetTimeout(2 * time.Second)

	return s, nil
}

func (s *Server) Serve() error {
	log.Infof("TFTP server listening on: %s", s.Address)
	return s.srv.ListenAndServe(s.Address)
}

func (s *Server) Shutdown() {
	s.srv.Shutdown()
}
