// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tftp

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pin/tftp/v3"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/provision"
	"github.com/ubccr/grendel/pkg/model"
)

func (s *Server) sendFile(fileName string, rf io.ReaderFrom) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Errorf("Failed to open %s: %s", fileName, err)
		return err
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		log.Errorf("Failed to send %s via tftp: %s", fileName, err)
		return err
	}

	log.Infof("Sent %s via tftp: %d bytes sent", fileName, n)
	return nil
}

func (s *Server) imageFileHandler(filePath string, rf io.ReaderFrom) error {
	imageName, fileType := filepath.Split(filePath)
	bootImage, err := s.DB.LoadBootImage(strings.TrimSuffix(imageName, "/"))
	if err != nil {
		log.Errorf("File not found: %s", filePath)
		return err
	}

	// TFTP requests carry no host context so only image level data is available to path templates
	data := map[string]any{
		"bootimage":  bootImage,
		"extra_vars": viper.GetStringMap("provision.extra_vars"),
	}

	switch {
	case fileType == "kernel":
		kernel, err := provision.RenderString("kernel", bootImage.KernelPath, data)
		if err != nil {
			log.Errorf("Failed to render kernel path for image %s: %s", bootImage.Name, err)
			return err
		}
		return s.sendFile(kernel, rf)
	case strings.HasPrefix(fileType, "initrd-"):
		i, err := strconv.Atoi(fileType[7:])
		if err != nil || i < 0 || i >= len(bootImage.InitrdPaths) {
			return fmt.Errorf("no initrd with ID %d", i)
		}
		initrd, err := provision.RenderString(fileType, bootImage.InitrdPaths[i], data)
		if err != nil {
			log.Errorf("Failed to render initrd path for image %s: %s", bootImage.Name, err)
			return err
		}
		return s.sendFile(initrd, rf)
	}

	return fmt.Errorf("File not found: %s", filePath)
}

func (s *Server) ReadHandler(token string, rf io.ReaderFrom) error {
	fwtype, err := model.ParseFirmwareToken(token)
	if err != nil {
		return s.imageFileHandler(token, rf)
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
