package tftp

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pin/tftp"
	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/model"
)

func (s *Server) ReadHandler(token string, rf io.ReaderFrom) error {
	log.Infof("Got TFTP read request with token: %s", token)

	fwtype, err := model.ParseFirmwareToken(token)
	if err != nil {
		log.Errorf("TFTP failed to parse token: %s", err)
		return err
	}

	bs, ok := s.FirmwareBin[fwtype]
	if !ok {
		log.Errorf("TFTP: unknown firmware type %d", fwtype)
		return fmt.Errorf("unknown firmware type %d", fwtype)
	}

	rf.(tftp.OutgoingTransfer).SetSize(int64(len(bs)))
	n, err := rf.ReadFrom(bytes.NewBuffer(bs))
	if err != nil && !strings.Contains(err.Error(), "User aborted") {
		log.Errorf("Failed to send firmware via tftp: %s", err)
		return err
	}

	log.Infof("Sent firmware %d via tftp: %d bytes sent", fwtype, n)

	return nil
}
