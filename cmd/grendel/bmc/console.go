// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"sync/atomic"

	"github.com/coder/websocket"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"golang.org/x/term"
)

// match the close codes in internal/api/bmc_console.go
const (
	consoleBusy   websocket.StatusCode = 4409
	consoleFailed websocket.StatusCode = 4500
)

var (
	consoleForce bool

	consoleCmd = &cobra.Command{
		Use:   "console <node>",
		Short: "Attach to a node's serial console over IPMI SOL",
		Long: `Attach to a node's serial console. The Grendel server opens an IPMI Serial over LAN session with the BMC and relays it, so the BMC credentials stay on the server.

Type ~. at the start of a line to disconnect, ~~ to send a ~, and ~? for help. A BMC allows one console session at a time, --force ends whoever else has it.

The BMC needs IPMI over LAN and SOL enabled, and the node needs serial redirection in its BIOS and a console= kernel argument to print anything.`,
		Example: `  grendel bmc console cpn-f01-04
  grendel bmc console cpn-f01-04 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return runConsole(command.Context(), args[0], consoleForce)
		},
	}
)

func init() {
	consoleCmd.Flags().BoolVar(&consoleForce, "force", false, "take the console over from another session")
	bmcCmd.AddCommand(consoleCmd)
}

func runConsole(ctx context.Context, node string, force bool) error {
	q := url.Values{"node": {node}}
	if force {
		q.Set("force", "true")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := apiclient.Websocket(ctx, "/v1/bmc/console?"+q.Encode())
	if err != nil {
		return err
	}
	defer conn.CloseNow()

	stdin := int(os.Stdin.Fd())
	if term.IsTerminal(stdin) {
		state, err := term.MakeRaw(stdin)
		if err != nil {
			return err
		}
		defer term.Restore(stdin, state)
	}

	// raw mode leaves newlines alone, so the messages carry their own carriage returns
	fmt.Fprintf(os.Stderr, "[connecting to %s]\r\n", node)

	var quit atomic.Bool
	closed := make(chan error, 1)
	go func() {
		for {
			typ, data, err := conn.Read(ctx)
			if err != nil {
				closed <- err
				return
			}
			if typ == websocket.MessageText {
				fmt.Fprintf(os.Stderr, "[connected to %s, ~. to disconnect, ~? for help]\r\n", node)
				continue
			}
			os.Stdout.Write(data)
		}
	}()

	go func() {
		f := &escapeFilter{lineStart: true}
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				out, action := f.filter(buf[:n])
				if len(out) > 0 {
					if werr := conn.Write(ctx, websocket.MessageBinary, out); werr != nil {
						return
					}
				}
				switch action {
				case escapeQuit:
					quit.Store(true)
					conn.Close(websocket.StatusNormalClosure, "")
					cancel()
					return
				case escapeHelp:
					fmt.Fprint(os.Stderr, "\r\n[~. disconnect, ~~ send ~, ~? this help]\r\n")
				}
			}
			// with stdin at its end, from a pipe or /dev/null, keep showing the console until the server closes it
			if err != nil {
				return
			}
		}
	}()

	err = <-closed
	fmt.Fprint(os.Stderr, "\r\n")

	switch status := websocket.CloseStatus(err); {
	case quit.Load(), errors.Is(err, context.Canceled), status == websocket.StatusNormalClosure:
		fmt.Fprintf(os.Stderr, "[disconnected from %s]\r\n", node)
		return nil
	case status == consoleBusy:
		return fmt.Errorf("%s: %s", node, closeText(err))
	case status == consoleFailed:
		return fmt.Errorf("console on %s: %s", node, closeText(err))
	case errors.Is(err, io.EOF):
		return fmt.Errorf("console on %s: connection to the server lost", node)
	default:
		return fmt.Errorf("console on %s: %w", node, err)
	}
}

// closeText is the reason the server gave in its close frame
func closeText(err error) string {
	var ce websocket.CloseError
	if errors.As(err, &ce) && ce.Reason != "" {
		return ce.Reason
	}
	return err.Error()
}

type escapeAction int

const (
	escapeNone escapeAction = iota
	escapeQuit
	escapeHelp
)

// escapeFilter picks the ssh style ~ escapes out of keyboard input. They only count at the start of a line, so a ~ in a path or a command still reaches the console
type escapeFilter struct {
	lineStart bool
	pending   bool
}

// filter returns the bytes to send to the console, and stops at an escape that needs acting on
func (f *escapeFilter) filter(in []byte) ([]byte, escapeAction) {
	out := make([]byte, 0, len(in))
	for _, b := range in {
		if f.pending {
			f.pending = false
			switch b {
			case '.':
				return out, escapeQuit
			case '?':
				return out, escapeHelp
			case '~':
				out = append(out, '~')
				f.lineStart = false
				continue
			default:
				out = append(out, '~')
			}
		} else if f.lineStart && b == '~' {
			f.pending = true
			continue
		}

		out = append(out, b)
		f.lineStart = b == '\r' || b == '\n'
	}
	return out, escapeNone
}
