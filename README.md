# Grendel - Bare Metal Provisioning for HPC

Grendel is a fast, easy to use bare metal provisioning system for High
Performance Computing (HPC) Linux clusters. Grendel simplifies the deployment
and administration of physical compute clusters both large and small. It's
developed by the University at Buffalo Center for Computational Research (CCR)
with more than 20 years of experience in HPC. Grendel is under active
development and currently runs CCR's production HPC clusters ranging from 200
to 1500 nodes.

## Key Features

* DHCP/PXE/TFTP provisioning
* DNS forward and reverse resolution
* Automatic host discovery
* Diskful and Stateless (Live image) provisioning
* BMC/iDRAC control via RedFish and IPMI
* Authorized provisioning using JWT tokens
* Rest API
* Easy installation (single binary with no deps)

## Hacking

Building Grendel requires Go v1.13 or greater:

```
$ git clone https://github.com/ubccr/grendel
$ cd grendel
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
  serve       Run services

Flags:
  -c, --config string   config file
      --debug           Enable debug messages
  -h, --help            help for grendel
      --verbose         Enable verbose messages

Use "grendel [command] --help" for more information about a command.
```

## License

Grendel is released under the GPLv3 license. See the LICENSE file.
