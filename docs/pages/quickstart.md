# Quickstart

Getting started with Grendel is easy. This quickstart guide assumes you have a
rack of servers connected to a switch and configured to PXE boot. We also
assume the server in which you're installing Grendel is either on the same
subnet as the nodes or the switch is setup to relay DHCP packets to your
Grendel server.

!!! note
    This guide shows how to get up and running with Grendel quickly. For
    notes on production deployments please see [here](/production).

## Installation

To install Grendel download a copy of the binary here.

```
$ ./grendel --version
```

## Configuration

Grendel can be configured using a `TOML` file. See
[here](https://github.com/ubccr/grendel/blob/master/grendel.toml.sample) for a
sample.

## Assemble the Boot Image

A boot image defines the Linux Kernel, Initrd files and optionally any command
line arguments. Boot images are defined using a simple JSON format. Grendel can
boot any Linux kernel and initrd, for example [Flatcar](https://www.flatcar-linux.org/) 
Linux is a distro specifically designed for running containers (fork of
CoreOS). Grab a copy of the Flatcar kernel and initrd here:

```
$ wget http://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_pxe_image.cpio.gz
$ wget http://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_pxe.vmlinuz
```

The corresponding boot image config for Grendel would look like this:

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

Grendel also supports booting LiveOS images. These are Linux kernel/initrd
images built with [Dracut live](https://mirrors.edge.kernel.org/pub/linux/utils/boot/dracut/dracut.html#_booting_live_images) 
support. For more details on building custom Linux images see our [Grendel Images](https://github.com/ubccr/grendel-images)
project. We also provide pre-built images for use in testing Grendel. You can
download them [here](https://github.com/ubccr/grendel-images/releases). 

For this tutorial, we'll be using the pre-built Ubuntu 20.04 image. First,
download the kernel, initramfs, and squashfs files:

```
$ wget https://github.com/ubccr/grendel-images/releases/download/build-2020-07-04/ubuntu-focal-vmlinuz
$ wget https://github.com/ubccr/grendel-images/releases/download/build-2020-07-04/ubuntu-focal-initramfs.img
$ wget https://github.com/ubccr/grendel-images/releases/download/build-2020-07-04/ubuntu-focal-squashfs.img
```

Then create the following JSON file `boot-image.json`:

```json
[{
    "name": "ubuntu-focal-live",
    "kernel": "ubuntu-focal-vmlinuz",
    "initrd": [
        "ubuntu-focal-initramfs.img"
    ],
    "liveimg": "ubuntu-focal-squashfs.img",
    "cmdline": "root=live:$liveimg BOOTIF=$mac rd.neednet=1 ip=dhcp"
}]
```

!!! warning
    Do not use these pre-built images in production. They are for testing purposes only.
    Default root password is: `ilovelinux`

## Importing Nodes

Nodes can be imported into Grendel from a file in a simple JSON format:

```json
[{
    "name": "cpn-d13-18",
    "provision": true,
    "boot_image": "ubuntu-focal-live",
    "interfaces": [
        {
            "fqdn": "cpn-d13-18.compute.ccr.buffalo.edu",
            "ip": "10.64.25.36",
            "mac": "41:69:AE:E6:61:06"
        }
    ]
}]
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
    be booted into Ubuntu

```
sudo ./grendel --verbose serve --hosts hosts.json --images boot-image.json
```
