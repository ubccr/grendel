# Grendel Changelog

## [Unreleased]

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

[Unreleased]: https://github.com/ubccr/grendel/compare/v0.0.5...HEAD
[0.0.1]: https://github.com/ubccr/grendel/releases/tag/v0.0.1
[0.0.2]: https://github.com/ubccr/grendel/releases/tag/v0.0.2
[0.0.4]: https://github.com/ubccr/grendel/releases/tag/v0.0.4
[0.0.5]: https://github.com/ubccr/grendel/releases/tag/v0.0.5
