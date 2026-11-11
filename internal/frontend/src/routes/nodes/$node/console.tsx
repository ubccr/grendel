import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { createFileRoute } from "@tanstack/react-router";
import { Eraser, Plug, PlugZap, Unplug } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

export const Route = createFileRoute("/nodes/$node/console")({
  component: RouteComponent,
});

// match the close codes in internal/api/bmc_console.go
const CONSOLE_BUSY = 4409;
const CONSOLE_ATTACHED = "attached";

type Status =
  | { state: "idle" }
  | { state: "connecting" }
  | { state: "connected" }
  | { state: "closed"; reason: string; busy: boolean; clean: boolean };

function RouteComponent() {
  const { node } = Route.useParams();
  const holder = useRef<HTMLDivElement>(null);
  const term = useRef<Terminal | null>(null);
  const socket = useRef<WebSocket | null>(null);
  const [status, setStatus] = useState<Status>({ state: "idle" });

  useEffect(() => {
    if (!holder.current) return;
    const t = new Terminal({
      convertEol: false,
      cursorBlink: true,
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
      fontSize: 13,
      scrollback: 5000,
      theme: { background: "#0a0a0a" },
    });
    const fit = new FitAddon();
    t.loadAddon(fit);
    t.open(holder.current);
    fit.fit();
    term.current = t;

    const observer = new ResizeObserver(() => fit.fit());
    observer.observe(holder.current);

    const encoder = new TextEncoder();
    const input = t.onData((d) => {
      if (socket.current?.readyState === WebSocket.OPEN) socket.current.send(encoder.encode(d));
    });

    return () => {
      input.dispose();
      observer.disconnect();
      socket.current?.close(1000);
      t.dispose();
    };
  }, []);

  const connect = useCallback(
    (force: boolean) => {
      socket.current?.close(1000);
      const scheme = window.location.protocol === "https:" ? "wss" : "ws";
      const query = new URLSearchParams({ node });
      if (force) query.set("force", "true");
      const ws = new WebSocket(`${scheme}://${window.location.host}/v1/bmc/console?${query}`);
      ws.binaryType = "arraybuffer";
      socket.current = ws;
      setStatus({ state: "connecting" });

      ws.onmessage = (e) => {
        if (typeof e.data === "string") {
          if (e.data === CONSOLE_ATTACHED) {
            setStatus({ state: "connected" });
            term.current?.focus();
          }
          return;
        }
        term.current?.write(new Uint8Array(e.data));
      };
      ws.onclose = (e) => {
        if (socket.current !== ws) return;
        socket.current = null;
        setStatus({
          state: "closed",
          reason: e.reason || (e.code === 1006 ? "connection to the server lost" : ""),
          busy: e.code === CONSOLE_BUSY,
          clean: e.code === 1000,
        });
      };
    },
    [node],
  );

  const disconnect = () => socket.current?.close(1000);
  const open = status.state === "connecting" || status.state === "connected";

  return (
    <div className="my-4 grid gap-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <StatusBadge status={status} />
          {status.state === "closed" && status.reason && (
            <span className="truncate text-sm text-muted-foreground">{status.reason}</span>
          )}
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" size="sm" onClick={() => term.current?.clear()}>
            <Eraser />
            Clear
          </Button>
          {open ? (
            <Button variant="secondary" size="sm" onClick={disconnect}>
              <Unplug />
              Disconnect
            </Button>
          ) : status.state === "closed" && status.busy ? (
            <Button size="sm" onClick={() => connect(true)}>
              <PlugZap />
              Take over
            </Button>
          ) : (
            <Button size="sm" onClick={() => connect(false)}>
              <Plug />
              Connect
            </Button>
          )}
        </div>
      </div>

      {status.state === "closed" && status.busy && (
        <p className="text-sm text-muted-foreground">
          Someone else has this console open. Take over ends their session.
        </p>
      )}

      <div className="overflow-hidden rounded-lg border bg-[#0a0a0a] p-2">
        <div ref={holder} className="h-[32rem] w-full" />
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: Status }) {
  switch (status.state) {
    case "idle":
      return <Badge variant="secondary">not connected</Badge>;
    case "connecting":
      return <Badge variant="secondary">connecting…</Badge>;
    case "connected":
      return <Badge className="bg-emerald-600 hover:bg-emerald-600">connected</Badge>;
    case "closed":
      return status.clean ? (
        <Badge variant="secondary">disconnected</Badge>
      ) : (
        <Badge variant="destructive">{status.busy ? "in use" : "closed"}</Badge>
      );
  }
}
