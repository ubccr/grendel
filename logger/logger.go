// This code was adoped from https://github.com/coredhcp/coredhcp/blob/master/logger/logger.go
//
// MIT License
//
// Copyright (c) 2018 coredhcp
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package logger

import (
	"sync"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	log_prefixed "github.com/x-cray/logrus-prefixed-formatter"
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
		logger.SetFormatter(&log_prefixed.TextFormatter{
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
