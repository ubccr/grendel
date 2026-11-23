// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"context"
	"errors"
	"fmt"

	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/pkg/model"
)

var log = logger.GetLogger("SWITCH")

const (
	NOSArista   = "arista"
	NOSSonic    = "sonic"
	NOSDellOS10 = "dellos10"
)

var (
	ErrUnknownNOS    = errors.New("failed to determine switch NOS")
	ErrNoCredentials = errors.New("failed to get switch credentials from config file")
	ErrNoInterfaces  = errors.New("switch has no interfaces")
)

// Codes name why a switch couldn't be read, for clients that explain how to fix each case. Unlike the messages they are part of the API and must not change
const (
	CodeUnknownNOS    = "unknown_nos"
	CodeNoCredentials = "no_credentials"
	CodeNoInterfaces  = "no_interfaces"
	CodeUnreachable   = "unreachable"
)

// ErrorCode is the code for an error from New or a driver call; anything not a setup problem is the switch not answering as expected
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrUnknownNOS):
		return CodeUnknownNOS
	case errors.Is(err, ErrNoCredentials):
		return CodeNoCredentials
	case errors.Is(err, ErrNoInterfaces):
		return CodeNoInterfaces
	}
	return CodeUnreachable
}

// NetworkSwitch reuses one connection across its calls, so a caller makes all of them on one value and then closes it
type NetworkSwitch interface {
	// GetInterfaceDescriptions returns every interface by name, with an empty description for those that have none
	GetInterfaceDescriptions(ctx context.Context) (map[string]string, error)
	GetMACTable(ctx context.Context) (model.MACTable, error)
	GetLLDPNeighbors(ctx context.Context) (model.LLDPNeighbors, error)
	Close() error
}

func HostNOS(host *model.Host) string {
	for _, nos := range []string{NOSArista, NOSSonic, NOSDellOS10} {
		if host.HasTags(nos) {
			return nos
		}
	}
	return ""
}

func New(host *model.Host, cfg Config) (NetworkSwitch, error) {
	nos := HostNOS(host)
	if nos == "" {
		return nil, ErrUnknownNOS
	}
	// TODO: use the management interface once interfaces have a type. The bmc flag cannot mark it, because it also decides which hosts the bmc commands contact
	if len(host.Interfaces) == 0 {
		return nil, ErrNoInterfaces
	}

	addr := host.Interfaces[0].AddrString()
	// an empty address would be dialed as ":port", which is this host
	if addr == "" {
		return nil, fmt.Errorf("%w with an address: its first interface has no IP", ErrNoInterfaces)
	}

	return NewFromAddr(addr, nos, cfg)
}

func NewFromAddr(addr, nos string, cfg Config) (NetworkSwitch, error) {
	switch nos {
	case NOSArista, NOSSonic:
		if cfg.Username == "" || cfg.Password == "" {
			log.Warn("Set switch.username and switch.password in grendel.toml to query switches over gNMI")
			return nil, ErrNoCredentials
		}
		return NewGNMI(addr, nos, cfg), nil
	case NOSDellOS10:
		return NewSNMP(addr, cfg)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownNOS, nos)
	}
}
