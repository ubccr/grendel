// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package tftp

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pin/tftp"
	"github.com/ubccr/grendel/model"
)

func (s *Server) ReadHandler(token string, rf io.ReaderFrom) error {
	fwtype, err := model.ParseFirmwareToken(token)
	if err != nil {
		log.Errorf("failed to parse token: %s", err)
		return err
	}

	log.Infof("Got read request for firmware type: %d", fwtype)

	bs := fwtype.ToBytes()
	if bs == nil {
		log.Errorf("Failed to fetch firmware %d: %s", fwtype, err)
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
