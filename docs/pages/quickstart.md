# Quickstart

Getting started with Grendel is easy. This quickstart guide assumes you have a
rack of servers connected to a switch and configured to PXE boot. We also
assume the server in which you're installing Grendel is either on the same
subnet as the nodes or the switch is setup to relay DHCP packets to your
Grendel server.

!!! note
    This guide shows how to get up and running with Grendel quickly. For
    production deployments please see the RPM installation guide.

## Installation

To install Grendel download a copy of the binary here.

```
$ ./grendel --version
```

## Configuration

Grendel can be configured using a `TOML` file. 

## Assemble the Boot Image

A boot image defines the Linux Kernel, Initrd files and optionally any command
line arguments. Boot images are defined using a simple JSON format. For this
tutorial we'll be using Flatcar container linux. To grab a copy of the Flatcar
kernel and initrd run the following commands:

```
$ wget http://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_pxe_image.cpio.gz
$ wget http://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_pxe.vmlinuz
```

The create the following JSON file `boot-image.json`:

```json
[{
    "name": "flatcar",
    "kernel": "flatcar_production_pxe.vmlinuz",
    "initrd": [
        "flatcar_production_pxe_image.cpio.gz"
    ],
    "cmdline": "flatcar.autologin"
}]
```

## Importing Nodes

Nodes can be imported into Grendel from a file in a simple JSON format. At
minimum a node requires a name, MAC address, IP address and FQDN:

```json
{
    "name": "cpn-d13-18",
    "provision": true,
    "interfaces": [
        {
            "fqdn": "cpn-d13-18.compute.ccr.buffalo.edu",
            "ip": "10.64.25.36",
            "mac": "41:69:AE:E6:61:06"
        }
    ]
}
```

You can use any process you wish for assembling your nodes into this format. To
make this easier, Grendel comes with a few built-in ways to auto-discover nodes
from querying a switch, snooping DHCP packets, parsing a `dhcpd.leases` file,
or importing from a simple `TSV` file. As a simple example, suppose you've
harvested the required information for each of your nodes in a TAB separated
file `hosts.tsv`:

```
name         mac                 ip             fqdn
cpn-d13-18   41:69:AE:E6:61:06   10.64.25.36    cpn-d13-18.compute.ccr.buffalo.edu
cpn-d13-19   41:69:AE:E6:71:16   10.64.25.35    cpn-d13-19.compute.ccr.buffalo.edu
cpn-d13-20   41:69:AE:E6:81:02   10.64.25.34    cpn-d13-20.compute.ccr.buffalo.edu
```

We can convert this into Grendel's host JSON format using the following
command:

```
./grendel discover file --input hosts.tsv > hosts.json
```

## Start services

Now that we've defined our hosts and a boot image we can start all the services
necessary to netboot the nodes by running Grendel:

!!! warning
    If your nodes are actively sending DHCP PXE boot requests these nodes will
    be booted into Flatcar

```
sudo ./grendel --verbose serve --hosts hosts.json --images boot-image.json
```
