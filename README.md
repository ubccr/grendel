![Grendel](docs/pages/images/logo-lg.png)

# Grendel - Bare Metal Provisioning for HPC

[![Documentation Status](https://readthedocs.org/projects/grendel/badge/?version=latest)](https://grendel.readthedocs.io/en/latest/?badge=latest)

Grendel is a fast, easy to use bare metal provisioning system for High
Performance Computing (HPC) Linux clusters. Grendel simplifies the deployment
and administration of physical compute clusters both large and small. It's
developed by the University at Buffalo Center for Computational Research (CCR)
with more than 20 years of experience in HPC. Grendel is under active
development and currently runs CCR's production HPC clusters ranging from 200
to 1500 nodes.

## Key Features

* Web based frontend
  * [Frontend demo](https://grendel-demo.ccr.buffalo.edu) User/pass = `grendel-demo`
* DHCP/PXE/TFTP provisioning
* DNS forward and reverse resolution
* Automatic host discovery
* Diskful and Stateless provisioning
* BMC/iDRAC control via RedFish and IPMI
* Authorized provisioning using [Branca](https://branca.io/) tokens
* Rest API
* Easy installation (single binary with no deps)

## Project status

Grendel is under heavy development and any API's will likely change
considerably before a more stable release is available. Use at your own risk.
Feedback and pull requests are more than welcome!

## Quickstart with QEMU/KVM

The following steps will show how to PXE boot a linux virtual machine using
QEMU/KVM and install Flatcar linux using Grendel. A demo of installing and using 
Grendel can be found [here](https://asciinema.org/a/b17YFY2uTpRotvUnXx8TVXx39).

### Installation

To install Grendel download a copy of the binary [here](https://github.com/ubccr/grendel/releases).

```
$ tar xvzf grendel-0.x.x-linux-amd64.tar.gz
$ cd grendel-0.x.x-linux-amd64/
$ ./grendel --help
```

### Create a TAP device

```
$ sudo ip tuntap add name tap0 mode tap user ${LOGNAME}
$ sudo ip addr add 192.168.10.254/24 dev tap0
$ sudo ip link set up dev tap0
```

For RedHat/CentOS

```
$ sudo firewall-cmd --zone=trusted --change-interface=tap0
```

For Debian/Ubuntu

```
$ sudo ufw allow in on tap0
```


### Create a boot Image file

```
$ wget http://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_pxe_image.cpio.gz
$ wget http://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_pxe.vmlinuz
```

Create the following JSON file `image.json`:

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

### Create a host file

Create the following JSON file `host.json`:

```json
[{
    "name": "tux01",
    "provision": true,
    "boot_image": "flatcar",
    "interfaces": [
        {
            "fqdn": "tux01.localhost",
            "ip": "192.168.10.12/24",
            "mac": "DE:AD:BE:EF:12:8C"
        }
    ]
}]
```

### Start Grendel services

```
$ sudo ./grendel --verbose serve --hosts host.json --images image.json --listen 192.168.10.254
```

Note: The serve command requires root privileges to bind to lower level ports.
If you don't want to run as root you can allow Grendel to bind to privileged
with the following command:

```
$ sudo setcap CAP_NET_BIND_SERVICE,CAP_NET_RAW=+eip /path/to/grendel
```

### PXE Boot the linux virtual machine

In another terminal window run the following commands:

```
$ qemu-system-x86_64 -m 2048 -boot n -device e1000,netdev=net0,mac=DE:AD:BE:EF:12:8C -netdev tap,id=net0,ifname=tap0,script=no
```

## Hacking

Building Grendel requires Go v1.23 or greater. Building iPXE requires packages
lzma-sdk-devel, xz-devel, and gcc-aarch64-linux-gnu:

```
$ git clone --recursive https://github.com/ubccr/grendel
$ cd grendel/firmware
$ make build
$ make bindata
$ cd ..
$ go build .
$ ./grendel help
Bare Metal Provisioning for HPC

Usage:
  grendel [command]

Available Commands:
  bmc         Query BMC devices
  discover    Auto-discover commands
  help        Help about any command
  host        Host commands
  image       Boot Image commands
  serve       Run services

Flags:
  -c, --config string     config file
      --debug             Enable debug messages
      --endpoint string   Grendel API endpoint (default "grendel-api.socket")
  -h, --help              help for grendel
      --verbose           Enable verbose messages

Use "grendel [command] --help" for more information about a command.
```

## Publications

- Andrew E. Bruno, Salvatore J. Guercio, Doris Sajdak, Tony Kew, and Matthew D.
  Jones. 2020. Grendel: Bare Metal Provisioning System for High Performance
  Computing. In Practice and Experience in Advanced Research Computing (PEARC
  ’20). Association for Computing Machinery, New York, NY, USA, 13–18.
  DOI: https://doi.org/10.1145/3311790.3396637

- PEARC ’20 Paper Presentation:
  [video](https://www.acsu.buffalo.edu/~aebruno2/talks/grendel-pearc20.mp4) |
  [slides](https://www.acsu.buffalo.edu/~aebruno2/talks/grendel)


## Acknowledgments

PXE booting is based on [Pixiecore](https://github.com/danderson/netboot/tree/master/pixiecore) by Dave
Anderson. DHCP implementation makes heavy use of this excellent [packet library](https://github.com/insomniacslk/dhcp). 
DNS implementation uses [this library](https://github.com/miekg/dns). TFTP implementation uses [this
library](https://github.com/pin/tftp). Backend database runs [BuntDB](https://github.com/tidwall/buntdb) or [sqlite](https://github.com/mattn/go-sqlite3). 
NodeSet/RangeSet algorithms ported from [ClusterShell](https://github.com/cea-hpc/clustershell)

## License

Grendel is released under the GPLv3 license. See REUSE.toml
