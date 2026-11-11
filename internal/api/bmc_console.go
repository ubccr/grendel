// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/pkg/model"
)

const (
	// ConsoleBusy is the close code sent when another session holds the serial console, so a client can offer to take it over
	ConsoleBusy websocket.StatusCode = 4409
	// ConsoleFailed is the close code for a console that could not be opened or was lost, with the reason in the close frame
	ConsoleFailed websocket.StatusCode = 4500

	consolePingInterval = 30 * time.Second
)

// ConsoleAttached is the one text message the server sends, once the BMC has given the console to this session
const ConsoleAttached = "attached"

// BmcConsole upgrades to a websocket carrying a node's serial console over IPMI SOL. Binary messages are console bytes in both directions, a text message says the console is attached, and the close frame says why the console ended
//
// It is registered on the mux directly rather than through fuego, whose logging middleware wraps the response writer in a type that cannot be hijacked
func (h *Handler) BmcConsole(w http.ResponseWriter, r *http.Request) {
	node := r.URL.Query().Get("node")
	if node == "" {
		ErrorSerializer(w, r, badRequest(errors.New("no node given"), "specify node"))
		return
	}

	hosts, err := h.requireNodes(node)
	if err != nil {
		ErrorSerializer(w, r, err)
		return
	}
	if len(hosts) != 1 {
		ErrorSerializer(w, r, badRequest(fmt.Errorf("%d nodes selected", len(hosts)), "a console attaches to one node"))
		return
	}
	targets := redfish.TargetsFromHosts(hosts)
	if len(targets) == 0 {
		ErrorSerializer(w, r, notFound(node+" has no interface marked bmc"))
		return
	}
	target := targets[0]
	force := r.URL.Query().Get("force") == "true"

	// a console stays open far longer than the server's write timeout, and the hijacked connection keeps whatever deadline it had
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})
	_ = rc.SetReadDeadline(time.Time{})

	// Accept checks that the Origin matches the Host, which keeps another site from opening a console with the user's cookie
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Debugf("console upgrade failed for %s: %s", node, err)
		return
	}
	defer conn.CloseNow()

	// the request context belongs to a connection that has been hijacked, so the console is tied to the websocket instead
	ctx, cancel := context.WithCancel(context.WithoutCancel(r.Context()))
	defer cancel()

	in, inWriter := io.Pipe()
	go func() {
		defer cancel()
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				inWriter.CloseWithError(io.EOF)
				return
			}
			if _, err := inWriter.Write(data); err != nil {
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(consolePingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := conn.Ping(ctx); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	started := time.Now()
	attached := false
	err = h.Redfish.Console(ctx, target, in, &consoleWriter{ctx: ctx, conn: conn}, redfish.ConsoleOptions{
		Force: force,
		OnActive: func() {
			attached = true
			_ = conn.Write(ctx, websocket.MessageText, []byte(ConsoleAttached))
			msg := "bmc console opened on " + node
			if force {
				msg += ", taking it over"
			}
			h.writeEvent(ctx, model.SeverityInfo.String(), msg)
		},
	})
	if attached {
		h.writeEvent(ctx, model.SeverityInfo.String(), fmt.Sprintf("bmc console closed on %s after %s", node, time.Since(started).Round(time.Second)))
	}

	switch {
	case err == nil:
		conn.Close(websocket.StatusNormalClosure, "console closed")
	case errors.Is(err, redfish.ErrConsoleBusy):
		conn.Close(ConsoleBusy, closeReason(err))
	default:
		log.Debugf("console on %s ended: %s", node, err)
		conn.Close(ConsoleFailed, closeReason(err))
	}
}

// closeReason fits an error into a close frame, which holds at most 123 bytes
func closeReason(err error) string {
	reason := err.Error()
	if len(reason) > 123 {
		reason = reason[:120] + "..."
	}
	return reason
}

// consoleWriter sends each chunk of console output as one binary message
type consoleWriter struct {
	ctx  context.Context
	conn *websocket.Conn
}

func (w *consoleWriter) Write(p []byte) (int, error) {
	if err := w.conn.Write(w.ctx, websocket.MessageBinary, p); err != nil {
		return 0, err
	}
	return len(p), nil
}
