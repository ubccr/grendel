package tftp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/pin/tftp"
	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/model"
)

func (s *Server) extractInfo(path string) (net.HardwareAddr, int, error) {
	pathElements := strings.Split(path, "/")
	if len(pathElements) != 2 {
		return nil, 0, errors.New("not found")
	}

	mac, err := net.ParseMAC(pathElements[0])
	if err != nil {
		return nil, 0, fmt.Errorf("invalid MAC address %q", pathElements[0])
	}

	i, err := strconv.Atoi(pathElements[1])
	if err != nil {
		return nil, 0, errors.New("not found")
	}

	return mac, i, nil
}

func (s *Server) ReadHandler(filename string, rf io.ReaderFrom) error {
	log.Infof("Got TFTP read request for file: %s", filename)

	_, i, err := s.extractInfo(filename)
	if err != nil {
		log.Errorf("TFTP: unknown path %q", filename)
		return fmt.Errorf("unknown path %q", filename)
	}

	bs, ok := s.FirmwareBin[model.Firmware(i)]
	if !ok {
		log.Errorf("TFTP: unknown firmware type %d", i)
		return fmt.Errorf("unknown firmware type %d", i)
	}

	rf.(tftp.OutgoingTransfer).SetSize(int64(len(bs)))
	n, err := rf.ReadFrom(bytes.NewBuffer(bs))
	if err != nil {
		log.Errorf("Failed to send firmware via tftp: %s", err)
		return err
	}

	log.Infof("Sent firmware %d via tftp: %d bytes sent", i, n)

	return nil
}
