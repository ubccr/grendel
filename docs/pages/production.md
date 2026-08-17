# Production Deployment

The following are tips for deploying Grendel in a production environment.

## CLI

Grendel installs two binaries. `grendeld` is the server: it runs the
services and the node discovery commands. `grendel` is the command
line client for everything else — nodes, images, users, roles, BMCs — and reaches
a running server over the API, either the unix socket or a TCP endpoint.

Only `grendeld` has to live on the provisioning host. `grendel` needs nothing
but a path to the API, so it can be installed anywhere administrators work.

## Container image

Releases publish an image to `ubccr/grendel` containing both binaries.
The default command is `grendeld serve --verbose`, and the sample config is
installed at `/etc/grendel/grendel.toml`, which is one of the default search
paths, so the server comes up without any arguments:

```
docker run -d --name grendel \
    --network host \
    -v grendel-data:/var/lib/grendel \
    -v /etc/grendel/grendel.toml:/etc/grendel/grendel.toml \
    ubccr/grendel:latest
```

There is an equivalent compose file at `configs/docker-compose.yml`.

`--network host` is what makes DHCP and PXE work. Both answer broadcast traffic
that a bridge network never delivers to the container. If you only run the
services that listen on ordinary sockets you can publish ports instead:

```
docker run -d -p 8080:8080 ubccr/grendel serve api --verbose
```

All state lives under `/var/lib/grendel`, so mount a volume there to keep the
database, boot images and templates across restarts.

The `grendel` client ships in the same image and the API socket is created in
the server's working directory, so it works over `docker exec` with no extra
configuration (the container below is also named `grendel`):

```
docker exec grendel grendel node list
```

`grendel version` prints the client and server versions, reading the latter
over the API. It reads nothing out of the database and fails only when the
server cannot be reached, which makes it the health check the compose file uses.

The image is built `FROM scratch` and holds the two binaries, a CA bundle and
the sample config, nothing else. There is no shell in it, so `docker exec
grendel sh` will not work. Run the binaries directly as above.

The `Dockerfile` at the top of the repository is not a from-source build. It
assembles the image from binaries GoReleaser has already compiled, so build it
with GoReleaser rather than `docker build`:

```
goreleaser release --clean --snapshot
```

## Running a subset of services

`grendeld serve` by default runs every service: `api`, `dhcp`, `dns`, `provision`, `pxe` and `tftp`. If you only want to run a subset, you may pass them as args or through the `--services` flag:

```
grendeld serve dhcp dns tftp
```

The same set can be pinned in the config file, and any service named on the
command line overrides it:

```toml
services = ["dhcp", "dns", "tftp"]
```

Per-service flags such as `--dhcp-listen` work whether you run one service or all
of them.

## Database settings

Unless you installed an rpm or deb package, Grendel's database is stored
entirely in memory by default and is not written to disk. This means any
changes to the Grendel database will be lost next time Grendel is restarted. If
you manage all your boot images and compute nodes via JSON files, simply make
sure Grendel is started with the following options:

```
grendeld serve --hosts /path/to/hosts.json --images /path/to/images.json
```

This will ensure the hosts and boot images are loaded each time Grendel is
started. Any changes to those files will require restarting Grendel to take
effect.

Alternatively, Grendel can be configured to persist the database to disk by
setting this config param:

```toml
#
# Path database file. Defaults to ":memory:" which uses in-memory store. Change
# this to a filepath for persisent storage.
#
dbpath = "/var/lib/grendel/grendel.db"
```

Any changes to the Grendel database will be persisted between restarts.

## DNS Stub Resolver

Grendel is not a recursive DNS resolver. In production deployments it's
recommended to run behind an [unbound](https://nlnetlabs.nl/projects/unbound/about/)
stub-resolver (or similar). This delegates management of compute node
forward/reverse DNS entirely to Grendel while keeping existing DNS
infrastructure in place. Here's some example configs for setting up Grendel as
an unbound stub resolver:

```yaml
stub-zone:
    name: "compute.ccr.buffalo.edu."
    stub-addr: IP_OF_GRENDEL_SERVER

stub-zone:
    name: "65.10.in-addr.arpa."
    stub-addr: IP_OF_GRENDEL_SERVER

stub-zone:
    name: "129.10.in-addr.arpa."
    stub-addr: IP_OF_GRENDEL_SERVER
```

The above will delegate resolution of `compute.ccr.buffalo.edu` zone, along
with reverse resolution of `10.65.0.0/16` and `10.129.0.0/16` subnets to
Grendel. You will obviously need to replace those with the appropriate zones
for your network.

When adding compute nodes to grendel, you can set the FQDN of the host to be
part of the stub-zone like so:

```json
{
  "name": "cpn-d13-08",
  "interfaces": [
    {
      "fqdn": "bmc-d13-08.compute.ccr.buffalo.edu",
      "ip": "10.129.24.8/24",
      "bmc": true
    },
    {
      "fqdn": "cpn-d13-08.compute.ccr.buffalo.edu",
      "ip": "10.65.24.8/24",
      "bmc": false
    }
  ],
  "provision": true
}
```

## Systemd unit file

In production it's recommended to setup Grendel in systemd. Here's an example
systemd unit file that is shipped with the rpm and deb packages:

```ini
[Unit]
Description=grendel server
After=syslog.target network.target

[Service]
Type=simple
User=grendel
Group=grendel
WorkingDirectory=/var/lib/grendel
ExecStart=/usr/bin/grendeld serve --verbose -c /etc/grendel/grendel.toml
Restart=on-failure
CapabilityBoundingSet=CAP_NET_BIND_SERVICE CAP_NET_RAW
AmbientCapabilities=CAP_NET_BIND_SERVICE CAP_NET_RAW
StateDirectory=grendel
ConfigurationDirectory=grendel

[Install]
WantedBy=multi-user.target
```
