# Provision Status

Grendel tracks how far each node got in its last provision. Only the latest status is kept: a new network boot
starts it over.

```
grendel node status cpn-f01-[01-40]
```

The web UI shows the same thing in the **Status** column of the nodes table and in the provision tab on the node page, along with any failed units.

## States

| State | Set when |
|---|---|
| `pxe` | the node's firmware asks Grendel's DHCP for a network boot |
| `ipxe` | the node fetches its iPXE script |
| `config` | the booted OS fetches its ignition, kickstart or cloud-init config |
| `booting` | `grendel-beacon` reports that systemd is still starting units |
| `failed` | `grendel-beacon` reports failed units, or a script reports a failure |
| `completed` | `grendel-beacon` reports a clean boot, or something POSTs the `complete` endpoint |

The first four come from Grendel itself and need nothing on the node. To know how the OS came
up, run `grendel-beacon` in it.

Without the beacon, a node stays at `config` until something POSTs `{{ $.endpoints.CompleteURL }}`. To only mark the status as completed and leave the provision flag alone, add `?unprovision=false`:

```
curl -X POST "{{ $.endpoints.CompleteURL }}?unprovision=false"
```

## grendel-beacon

`grendel-beacon` is a small static Linux binary, built for amd64 and arm64. It talks to systemd
directly over `/run/systemd/private` (or the system bus), so it needs nothing else installed and
works on any systemd distro.

```
grendel-beacon watch [--complete] [--interval 10s] [--timeout 1h]
```

`watch` follows systemd until the boot finishes. It reports `booting` while units are starting,
then the result:
- `completed` when the system reaches `running`
- `failed` with the list of failed units when it ends up `degraded`
- `failed` when the boot hasn't finished by `--timeout`

With `--complete`, a clean boot also turns provisioning off for the node, just as the `complete`
endpoint does. The node status will flip to completed regardless of this flag so you may leave it off for nodes that network boot every time.

```
grendel-beacon report --state failed --message "raid setup failed"
grendel-beacon report --state completed --complete
```

`report` sends a single report, for scripts such as a kickstart `%post`.

Both commands read the node's status URL from `--url` or `GRENDEL_STATUS_URL`. In a provision
template that URL is `{{ $.endpoints.StatusURL }}`. Use `--insecure` (or `GRENDEL_INSECURE=true`)
when the provision server has a self-signed certificate.

### Facts

Each report can carry **facts**: named values describing the node, shown on its Provision tab and returned by `GET /v1/nodes/provision-status`. `watch` collects them for its first progress report, so a stuck boot still shows what it is running, and again for its final report, once the boot timings exist.

Built in:

| Facts | From |
|---|---|
| `os.*` | every key in `/etc/os-release`, lowercased: `os.id`, `os.version_id`, `os.pretty_name`, `os.image_version`... |
| `kernel.release` | the running kernel, as `uname -r` prints it |
| `boot.*` | systemd's boot timings, split as `systemd-analyze` does: `boot.firmware`, `boot.loader`, `boot.kernel`, `boot.initrd`, `boot.userspace`, `boot.total` |

**Site facts** come from drop-ins in `/etc/grendel/beacon.d` (`--facts-dir` to change it):
- an executable is run (10 second limit) and its output read
- any other file is read as it is

Both hold `key=value` lines, and each key is prefixed with the file's name, so `beacon.d/slurm` printing `version=24.05` reports `slurm.version`.

```sh
#!/bin/sh
# /etc/grendel/beacon.d/slurm
echo "version=$(sinfo --version | cut -d' ' -f2)"
```

Rules for drop-ins:
- **File names:** lowercase letters, digits, `-` and `_`. Anything else, such as `slurm.rpmsave`, is skipped.
- **No overriding:** a drop-in can't replace a built in fact.
- **Checking:** `grendel-beacon facts` prints what would be sent, which is the quickest way to check a new drop-in.

`report` sends only the facts given with `--fact key=value`, since an installer's `%post` would otherwise report the installer's os-release. Add `--collect` to include the built in and drop-in facts as well:

```
grendel-beacon report --state booting --fact site.raid_layout=raid10
```

**Limits:**
- **Keys:** lowercase letters, digits, `.`, `_` and `-`, up to 64 characters.
- **Values:** up to 1 KB.
- **Count:** a node keeps at most 200 facts.
- **Invalid facts:** dropped.

Each report's facts are merged into the ones already reported for that provision. A new network boot starts over.

### Served by Grendel

The deb and rpm packages install `grendel-beacon-amd64` and `grendel-beacon-arm64` in
`/usr/lib/grendel/beacon`. Otherwise download them from the release and put them in a
directory on the Grendel server. Then turn on the beacon in the config:

```toml
[provision.beacon]
enabled = true
dir = "/usr/lib/grendel/beacon"
# path = "/opt/bin/grendel-beacon"
# complete = false
```

Grendel serves the binaries at `/beacon/<arch>`. The default butane template then adds three
things to the ignition config:
- a download of the binary for the node's arch to `path`
- `/etc/grendel/beacon.env` with the status URL
- an enabled `grendel-beacon.service` that runs `watch`

`complete = true` adds `--complete` to that service.

### Baked into the image

Install `grendel-beacon` in the image yourself and set `enabled = true` with `dir` left empty. The
butane template then writes the env file and the unit but doesn't download anything; point `path`
at where the image has the binary.

For a custom template, or a provisioning method other than ignition, the unit needs the status
URL:

```ini
# /etc/systemd/system/grendel-beacon.service
[Unit]
Description=Report provision status to grendel
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
Environment=GRENDEL_STATUS_URL={{ $.endpoints.StatusURL }}
ExecStart=/usr/bin/grendel-beacon watch

[Install]
WantedBy=multi-user.target
```

In a kickstart, the unit can be written and enabled in `%post`, so that it reports how the
installed OS comes up on its first boot.

## Security

Reports go to `POST /boot/<token>/status`, using the same signed boot token as the other `/boot`
routes:
- **No forging:** the token is encrypted and authenticated with `provision.secret`, so it can't be
  forged or changed to name another node. A token only lets someone report on its own node, and
  unprovision it with `complete`.
- **Expiry:** the token expires after `provision.token_ttl` (1 hour by default). A provision that
  runs longer than that stops being able to report.
- **Source address:** a report must come from one of the node's own addresses, checked against
  the TCP connection rather than any forwarded header. Set `provision.status_check_source = false`
  when nodes reach Grendel through NAT or a proxy.
- **Finished provisions:** once a provision has finished, only the boot that finished it can report
  again. `watch` sends the node's boot ID; `report` doesn't, so a report from an installer doesn't
  lock out the installed OS.
- **Token exposure:** the token is in the node's ignition config and `/etc/grendel/beacon.env`
  (readable only by root).
