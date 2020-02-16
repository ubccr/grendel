# What is Grendel?

![Diagram](images/logo-lg.png)

Grendel is a fast, easy to use, and full featured bare metal provisioning
system for High Performance Computing (HPC) Linux clusters. Grendel simplifies
the deployment and administration of physical compute clusters both large and
small. It's developed by the University at Buffalo Center for Computational
Research (CCR) with more than 20 years of experience in HPC. Grendel is under
active development and currently runs CCR's production HPC clusters ranging
from 200 to 1500 nodes.

## Key Features

* DHCP/PXE/TFTP provisioning
* DNS forward and reverse resolution
* Automatic host discovery
* Diskful and Stateless (Live image) provisioning
* BMC/iDRAC control via RedFish and IPMI
* Authorized provisioning using JWT tokens
* Rest API
* Easy installation (single binary with no deps)

## Goals

Grendel had one goal: to be simple to use. Building an HPC Linux Cluster can be
quite tedious. Grendel aims to be a modern, easy to use provisioning system for
organizations just want to netboot a rack of servers without much fuss.
