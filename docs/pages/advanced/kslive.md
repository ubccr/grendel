# Kickstarting Live Images

Grendel can be configured to kickstart compute nodes from live images. Building
live images is out of the scope of this document and it's assumed you already
have a live image built. In order to kickstart we need to have the appropriate
installation media for the RHEL based distribution. In this example, we'll be
using CentOS.  If you already have a local CentOS mirror you can skip the next
section.  Otherwise you can use Grendel to serve the installation media over
HTTP(S). 

## Serve the installation media

Grendel can be configured to serve installation media. These files are fetched
by anaconda during the kickstart process. This includes the installation program
runtime image to be loaded specified by the `inst.stage2` kernel argument.
Create a local copy of the CentOS media and add the following to `grendel.toml`:

```toml
[provision]
repo_dir = "/repo"
```

The above simply configures Grendel as a file server, serving any files in that
directory over HTTP. 

!!! note 
    `repo_dir` can be the top level directory which may contain multiple
    different local mirrors (for example centos, epel, etc). If so, you can
    specify the correct path in the kernel command line below.

After starting Grendel services you should see the output below indicating
Grendel is now serving the `repo_dir` configured above:

```
INFO PROVISION: Using repo dir: /repo
```

## Create the Boot Image JSON

Add the path to the live image and the kernel command line arguments for
kickstarting. The example below uses Grendel to serve all the provision assets.
Grendel will substitute `$kickstart` with the full URL to the kickstart file
served by Grendel. `$repo` will be the base url of the `repo_dir` defined above
and served by Grendel. You don't have to use Grendel to serve the provision
assets and can simply adjust the kernel command line arguments to the
appropriate URLs for your situation.

```json
[{
    "name": "centos7",
    "kernel": "/repo/centos/7.7.1908/os/x86_64/images/pxeboot/vmlinuz",
    "initrd": [
        "/repo/centos/7.7.1908/os/x86_64/images/pxeboot/initrd.img"
    ],
    "cmdline": "ks=$kickstart network ksdevice=bootif ks.device=bootif inst.stage2=$repo/centos/7.7.1908/os/x86_64",
    "liveimg": "/images/compute-node.squashfs"
}]
```

With the above config Grendel will generate an iPXE script, for example:

```
#!ipxe
kernel --name kernel http://192.168.10.254/boot/file/kernel

initrd --name initrd0 http://192.168.10.254/boot/file/initrd-0
boot kernel initrd=initrd0 ks=http://192.168.10.254/boot/kickstart network ksdevice=bootif ks.device=bootif inst.stage2=http://192.168.10.254/repo/centos/7.7.1908/os/x86_64
```

Grendel also includes a basic kickstart template which uses a live image for the
installation source, for example:

```
install
liveimg --url="http://192.168.10.254/boot/file/liveimg

lang en_US.UTF-8
selinux --disabled
keyboard us
timezone --utc America/New_York
skipx

network --bootproto dhcp --hostname tux01.local --device=de:ad:be:ef:12:8c
```
