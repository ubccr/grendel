# Grendel Changelog

## [0.0.15] - 2024-11-22

- Fix: frontend - inventory export download
- Feat: frontend - reworked inventory export variables
- Feat: frontend - added inventory import page

## [0.0.14] - 2024-09-24

- Update default grendel.toml config
- Feat: provision - added custom PDU prometheus service discovery endpoint
- Feat: frontend - added Power page to display panel and circuit information. required pdu tags: panel:1,circuit:1-3
- Feat: frontend - rack page now renders 0u PDUs
- Feat: cli - added bmc firmware and job commands to allow firmware updates via the Redfish API
- Feat: cli - added auto-downloading of the Dell update catalog & firmware updates
- Fix: bmc - power command now functions correctly without an --override set

## [0.0.13] - 2024-08-26

- Update retryablehttp, gjson, ipxe deps
- Feat: frontend - added CSV export with GO template support
- Feat: tors - added SONiC queries
- Feat: tors - added Arista EOS queries
- Fix: frontend - prevent initial admin user from needing to relog after registration
- Fix: bmc - concurrent map writes on BMC queries
- Fix: nfpm - use updated sample toml file rather than .default file

## [0.0.12] - 2024-08-05

- Update echo, fiber deps
- Fix provision - tftp segfault when starting with bind port permission error
- Feat: provision - added "proxmox" tag DHCP 250 response code for Proxmox Automated Installs
- Fix: frontend - segfault when non nodeset hostname is added
- Fix: frontend - rack selection checkboxes not preserving state after submitting an action
- Fix: frontend - sqlite session storage
- Feat: fronend - added Nodes page
- Feat: frontend - added Bond support in Host page
- Feat: frontend - added status page

## [0.0.11] - 2024-02-27

- Add prometheus service discovery.
- Add support for bonded nics
- Add arm64 pxe booting 
- Add Dell SONiC ZTP
- Remove Dell OS9 BMP support

## [0.0.10] - 2024-01-15

- Fixed "Disabled" user role not respecting auth middleware
- Fixed bulk adding nodes other than cpn || srv not having the correct bmc fqdn prefixes
- Improved notifications UI
- Improved rack actions UI
- Improved user table UI
- Added import host by JSON in UI
- Overhauled BMC package
- Deprecated IPMI support
- Added ClearSel action
- Added host power options & overrides
- Added bmc power cycle
- Improved redfish error handling
- Improved import sys config settings
- Improved bmc CLI
- Added CLI autoconfigure command
- Updated Tailwind to v3.4.0
- Updated gofish to v0.15.0

## [0.0.9] - 2023-11-29

- Add Web UI (@jafurlan)
- Fix bug in reverse resolution with ip masks
- Fix bug in reverse resolution with multiple names
- Add tags to status output
- Add network segment status command
- Add support for Dell ONIE and iDRAC auto config
- Allow multiple FQDN comma separated
- Add support for NVIDIA/Mellanox zerotouch
- Add support for Arista zerotouch deployment (ZTD)

## [0.0.8] - 2023-02-27

- Add DHCP multi-interface support [#12](https://github.com/ubccr/grendel/issues/12)
- Add support for Dell Zero-touch deployment (ZTD) of switches. The necessary
  DHCP options are added if a host is tagged with "dellztd" or "dellbmp" (to
  support older FTOS switches).
- Add support for serving image kernel/initrd assets over tftp
- Add subnet config settings to define gateway, dns, search domains and mtu. If
  a host IP falls in the subnet grendel will inherit these settings when
  offering dhcp leases.
- Add MTU and VLAN parameters to network interface json definitions.
- Fix incorrect padding in nodesets [#20](https://github.com/ubccr/grendel/issues/20)
- Add support for custom named provision templates
- Add `admin_ssh_pubkeys` config option

### BREAKING CHANGES

- Add netmask prefix to IPs in json definitions. Changes the `NetInterface.IP`
  type from `net.IP` to `netip.Prefix` which allows us to capture both the IP
  address and the network prefix. The raw IP stored in the json now has the
  following format: `x.x.x.x/xx`. This is a breaking change and will require a
  manual dump/restore of the grendel database.
- Rename `dhcp.router` config option to `dhcp.gateway`
- Remove unused `provision.scheme` config option

## [0.0.7] - 2022-06-09

- Fix bug in marshalling interface names 
- Fix ubuntu autoinstall template
    - Set root password in Ubuntu autoinstall template
    - Add dhcp4 network configs for default nic
- Templatize the kernel command line for images
- Add support for translating butane configs to ignition
- Update ipxe to latest release
- Add rpm/deb package building

## [0.0.6] - 2021-10-25

- Add support for kickstarting both uefi and legacy bios nodes from the same
  kickstart template based on tags
- Clean up status command
- Randomize DNS server list in DHCP responses to distribute load
- Fix duplicate checking. Properly update hosts with same name
- Add delete host/image API and CLI
- Add check for duplicate host IDs
- Fix bug in delay and fanout
- Add support for cloud-init and Ubuntu autoinstall
- Add support for power/on off in CLI


## [0.0.5] - 2021-03-26

- Handle DHCPINFORM
- Add support for turning on/off logging for specific services
- Switch to using native go embed support in v1.16 for assets
- Add feature to allow boot images to set custom provision templates
- Add support for tagging hosts
- Refactor NodeSet api to support N-dim folding
- Add new status cli command for displaying basic status of nodes

## [0.0.4] - 2020-07-04

- Add boot image cli commands
- Add provision cli commands
- Add api endpoint cli flag

## [0.0.2] - 2020-07-02

- Allow variable interpolation in image command line
- Add support for codesigning boot images
- Switch to using Branca instead of JWT
- Require host to be set to provisioned in order to fetch boot assets. 
- Add endpoint to unprovision host
- Switch to using openapi
- Add api endpoints for boot images

## [0.0.1] - 2020-02-17

- Initial release

[0.0.1]: https://github.com/ubccr/grendel/releases/tag/v0.0.1
[0.0.2]: https://github.com/ubccr/grendel/releases/tag/v0.0.2
[0.0.4]: https://github.com/ubccr/grendel/releases/tag/v0.0.4
[0.0.5]: https://github.com/ubccr/grendel/releases/tag/v0.0.5
[0.0.6]: https://github.com/ubccr/grendel/releases/tag/v0.0.6
[0.0.7]: https://github.com/ubccr/grendel/releases/tag/v0.0.7
[0.0.8]: https://github.com/ubccr/grendel/releases/tag/v0.0.8
[0.0.9]: https://github.com/ubccr/grendel/releases/tag/v0.0.9
[0.0.10]: https://github.com/ubccr/grendel/releases/tag/v0.0.10
[0.0.11]: https://github.com/ubccr/grendel/releases/tag/v0.0.11
[0.0.12]: https://github.com/ubccr/grendel/releases/tag/v0.0.12
[0.0.13]: https://github.com/ubccr/grendel/releases/tag/v0.0.13
[0.0.14]: https://github.com/ubccr/grendel/releases/tag/v0.0.14
[0.0.15]: https://github.com/ubccr/grendel/releases/tag/v0.0.15
[Unreleased]: https://github.com/ubccr/grendel/compare/v0.0.15...HEAD
