# Production Deployment

The following are tips for deploying Grendel in a production environment.

## Database settings

Unless you installed an rpm or deb package, Grendel's database is stored
entirely in memory by default and is not written to disk. This means any
changes to the Grendel database will be lost next time Grendel is restarted. If
you manage all your boot images and compute nodes via JSON files, simply make
sure Grendel is started with the following options:

```
grendel serve --hosts /path/to/hosts.json --images /path/to/images.json
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
ExecStart=/usr/bin/grendel serve --verbose -c /etc/grendel/grendel.toml
Restart=on-failure
CapabilityBoundingSet=CAP_NET_BIND_SERVICE CAP_NET_RAW
AmbientCapabilities=CAP_NET_BIND_SERVICE CAP_NET_RAW
StateDirectory=grendel
ConfigurationDirectory=grendel

[Install]
WantedBy=multi-user.target
```
