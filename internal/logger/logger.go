// Copyright (c) 2018 coredhcp
// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package logger

import (
	"sync"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var (
	globalLogger   *logrus.Logger
	getLoggerMutex sync.Mutex
)

// GetLogger returns a configured logger instance
func GetLogger(prefix string) *logrus.Entry {
	if prefix == "" {
		prefix = "<no prefix>"
	}
	if globalLogger == nil {
		getLoggerMutex.Lock()
		defer getLoggerMutex.Unlock()
		logger := logrus.New()
		logger.SetFormatter(&TextFormatter{
			FullTimestamp: true,
		})
		globalLogger = logger
	}
	return globalLogger.WithField("prefix", prefix)
}

// WithFile logs to the specified file in addition to the existing output.
func WithFile(log *logrus.Entry, logfile string) {
	log.Logger.AddHook(lfshook.NewHook(logfile, &logrus.TextFormatter{}))
}
