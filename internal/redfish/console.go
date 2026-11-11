// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	ipmi "github.com/bougou/go-ipmi/pkg/client"
	"github.com/bougou/go-ipmi/pkg/command/transport"
	"github.com/bougou/go-ipmi/pkg/types"
)

const ipmiConnectTimeout = 15 * time.Second

// ErrConsoleBusy means another SOL session holds the console, which a BMC allows only one of
var ErrConsoleBusy = errors.New("the serial console is in use by another session, use force to take it over")

type ConsoleOptions struct {
	// Force ends a SOL session someone else has open rather than failing with ErrConsoleBusy
	Force bool
	// OnActive is called once the console is attached, before any output
	OnActive func()
}

// Console attaches to a node's serial console over IPMI SOL, copying in to the console and the console to out until ctx ends, in reaches EOF, or the BMC ends the session
func (s *Service) Console(ctx context.Context, t Target, in io.Reader, out io.Writer, opts ConsoleOptions) error {
	c, err := s.ipmiConnect(ctx, t)
	if err != nil {
		return err
	}
	defer func() {
		// ctx is usually done by now, and the session should still be closed so the BMC doesn't hold a slot until it times out
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.Close(closeCtx)
	}()

	if err := activateSOL(ctx, c, opts.Force); err != nil {
		return err
	}

	if opts.OnActive != nil {
		opts.OnActive()
	}

	err = c.SOLStream(ctx, in, out)
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, io.EOF) {
		// the bmc ended it, often because someone took the console over. Deactivating here would end their session, since any admin session may deactivate the payload, and closing ours is enough
		if errors.Is(err, net.ErrClosed) {
			return errors.New("the bmc ended the console session")
		}
		if strings.Contains(err.Error(), "SOL deactivation") {
			return errors.New("the console was taken over by another session, or ended by the bmc")
		}
		return err
	}

	// the viewer left while the console was still ours, so hand it back now rather than when the bmc notices the session is gone
	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = c.DeactivatePayload(closeCtx, &transport.DeactivatePayloadRequest{PayloadType: types.PayloadTypeSOL, PayloadInstance: 1})
	return nil
}

// ipmiConnect opens an RMCP+ session, trying each configured credential the way connect does for redfish
func (s *Service) ipmiConnect(ctx context.Context, t Target) (*ipmi.Client, error) {
	if t.Addr == "" {
		return nil, errors.New("bmc interface has no usable ip address")
	}
	if len(s.cfg.Credentials) == 0 {
		return nil, errors.New("no bmc credentials configured, set bmc.user and bmc.password")
	}

	suites := make([]types.CipherSuiteID, 0, len(s.cfg.IPMICipherSuites))
	for _, id := range s.cfg.IPMICipherSuites {
		suites = append(suites, types.CipherSuiteID(id))
	}

	var lastErr error
	for _, cred := range s.cfg.Credentials {
		c, err := ipmi.NewClient(t.Addr, s.cfg.IPMIPort, cred.User, cred.Password)
		if err != nil {
			return nil, err
		}
		c.WithInterface(ipmi.InterfaceLanplus)
		if len(suites) > 0 {
			c.WithCipherSuiteID(suites...)
		}

		// go-ipmi retries each exchange several times, which leaves an unreachable bmc hanging for most of a minute
		connectCtx, cancel := context.WithTimeout(ctx, ipmiConnectTimeout)
		err = c.Connect(connectCtx)
		cancel()
		if err == nil {
			return c, nil
		}
		lastErr = err
		// an unreachable bmc fails the same way for every credential, so only a rejected login moves on to the next pair
		if !isIPMIAuthFailure(err) {
			return nil, fmt.Errorf("ipmi connect: %w", err)
		}
		log.Debugf("ipmi credentials for user %s rejected by %s, trying the next pair", cred.User, t.Name)
	}

	return nil, fmt.Errorf("no configured credentials were accepted over ipmi (tried %s): %w", strings.Join(s.credentialUsers(), ", "), lastErr)
}

// isIPMIAuthFailure picks out a rejected login, which go-ipmi reports from the RAKP exchange either as an RMCP+ status from the bmc or as a key exchange code that doesn't match
func isIPMIAuthFailure(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if _, ok := errors.AsType[net.Error](err); ok {
		return false
	}
	if status, ok := errors.AsType[*types.RmcpStatusError](err); ok {
		switch status.StatusCode() {
		case types.RmcpStatusCodeUnauthorizedName, types.RmcpStatusCodeInvalidIntegrityCheckValue, types.RmcpStatusCodeUnauthorizedRoleOfPriLevel, types.RmcpStatusCodeInvalidRole:
			return true
		}
		return false
	}
	return strings.Contains(err.Error(), "rakp")
}

// activateSOL asks for an encrypted console first, so the console isn't sent in the clear when the cipher suite can encrypt it, and takes the console over when forced
func activateSOL(ctx context.Context, c *ipmi.Client, force bool) error {
	activate := func(secure bool) error {
		_, err := c.ActivatePayload(ctx, &transport.ActivatePayloadRequest{
			PayloadType:          types.PayloadTypeSOL,
			PayloadInstance:      1,
			EnableEncryption:     secure,
			EnableAuthentication: secure,
		})
		return err
	}

	err := activate(true)
	if completionCode(err) == types.CodeActivatePayloadCannotActivateWithEncryption {
		err = activate(false)
	}
	if completionCode(err) == types.CodeActivatePayloadAlreadyActive {
		if !force {
			return ErrConsoleBusy
		}
		if _, derr := c.DeactivatePayload(ctx, &transport.DeactivatePayloadRequest{PayloadType: types.PayloadTypeSOL, PayloadInstance: 1}); derr != nil {
			return fmt.Errorf("failed to end the other console session: %w", derr)
		}
		err = activate(true)
		if completionCode(err) == types.CodeActivatePayloadCannotActivateWithEncryption {
			err = activate(false)
		}
	}

	switch completionCode(err) {
	case 0:
	case types.CodeActivatePayloadTypeDisabled:
		return errors.New("serial over lan is disabled on the bmc, or for this user")
	case types.CodeActivatePayloadAlreadyActive, types.CodeActivatePayloadActivationLimitReached:
		return ErrConsoleBusy
	}
	if err != nil {
		return fmt.Errorf("failed to activate serial over lan: %w", err)
	}
	return nil
}

// completionCode is 0 for a nil error or one that isn't a completion code
func completionCode(err error) types.CompletionCode {
	if respErr, ok := errors.AsType[*types.ResponseError](err); ok {
		return respErr.CompletionCode()
	}
	return 0
}
