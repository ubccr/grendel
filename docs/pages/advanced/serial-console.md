# Serial Console

Grendel can attach to a node's serial console over IPMI Serial over LAN (SOL), to watch it boot or
log in when the network is down. The Grendel server opens the SOL session with the BMC and relays it
over a websocket, so the BMC credentials never leave the server.

```
grendel bmc console cpn-f01-04
```

In the web UI, open a node and pick the **Console** tab, or use **Serial Console** in a node's
actions.

A BMC allows one SOL session at a time. If someone else has the console, Grendel says so rather
than taking it; `--force` (or **Take over** in the web UI) ends their session and gives you the
console. Opening the console tab does not connect on its own, for the same reason.

In the CLI, type `~.` at the start of a line to disconnect, `~~` to send a `~`, and `~?` for help.
Ctrl-C is sent to the node. Each console opened and closed is recorded in the event log, without
anything typed or printed.

## Requirements

Several things have to line up for a console to show anything:

- **Network**: UDP 623 open from the Grendel server to the BMCs.
- **BMC**: IPMI over LAN and SOL enabled, for the user Grendel logs in as. On a Dell iDRAC these
  are the `IPMILan.1.Enable` and `IPMISOL.1.Enable` attributes. iDRAC uses the same account for
  Redfish and IPMI, so `bmc.user` and `bmc.password` cover both.
- **BIOS**: serial redirection on, so the firmware and boot loader print to the serial port. On a
  Dell, set `SerialComm` to `OnConRedirCom2` (`grendel bmc bios set`), which puts the console on
  COM2, where iDRAC's SOL listens.
- **OS**: the kernel told to use the port, for example `console=tty0 console=ttyS1,115200n8` in the
  boot image's command line. The last `console=` is where most messages go, and where a login
  prompt appears if getty is enabled on it.

## Configuration

```toml
[bmc.ipmi]
# UDP port the BMCs listen on for IPMI
port = 623
# RMCP+ cipher suites to try, in order. 17 (SHA256, AES) is the iDRAC9 default, 3 (SHA1, AES)
# covers older BMCs
cipher_suites = [17, 3]
```

Credentials are the same ones the Redfish commands use, `bmc.user`/`bmc.password` followed by any
`[[bmc.defaults.credentials]]`, tried in order. The console is encrypted whenever the cipher suite
supports it.

## Access

The console is served at `GET /v1/bmc/console?node=<name>[&force=true]`, a websocket. Binary
messages carry console bytes in both directions, a text message `attached` arrives once the BMC has
given the console to the session, and the close frame carries the reason it ended: code `4409` when
the console is in use, `4500` when it could not be opened or was lost.

The built in `admin` and `user` roles can open consoles, `read-only` cannot, since a console takes
keystrokes.
