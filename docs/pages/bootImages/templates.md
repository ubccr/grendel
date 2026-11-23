# Auto Provisioning

While you can provision a node into a live installer, and finish the install with a keyboard and mouse, repeating this process for thousands of nodes is not ideal. This process is distro specific:

| Distro    | Method       |
|-----------|--------------|
| Debian    | [Preseed](https://wiki.debian.org/DebianInstaller/Preseed)                                            |
| Ubuntu    | [Autoinstall](https://canonical-subiquity.readthedocs-hosted.com/en/latest/intro-to-autoinstall.html) |
| Rocky     | [Anaconda](https://anaconda-installer.readthedocs.io/en/latest/kickstart.html)                        |
| RHEL      | [Anaconda](https://anaconda-installer.readthedocs.io/en/latest/kickstart.html)                        |

## Provision Templates

Kickstart and other provision templates are rendered through Go text templates. You can find some simple example templates [here](https://github.com/ubccr/grendel/tree/main/internal/render/templates). In these templates, you can call any package functions defined in the text/template package as well as what Grendel adds below:

- [Functions](https://github.com/ubccr/grendel/blob/main/internal/render/funcs.go)
- [Variables](https://github.com/ubccr/grendel/blob/main/internal/render/data.go)

### Checking templates

After editing a template, print what a node would be sent without booting it:

```
grendel node render cpn-01 kickstart
grendel node render cpn-[01-04] ipxe --arch aarch64
grendel node render cpn-01 ipxe --arch all
```

The template is `ipxe`, `kickstart`, `user_data`, `meta_data`, `butane`, `ignition` (the Butane template translated to Ignition), `proxmox`, a key of the image's `provision_templates`, or a template file name. Templates are read from disk on each run, and one that does not parse is reported. `--arch` stands in for the arch detected during a PXE boot on nodes with no arch set, and `--arch all` renders such a node once as `x86_64` and once as `aarch64`. The output holds a working boot token, and whatever secrets the template renders.

To check that every node can be served its boot image, run:

```
grendel node check all --failed
grendel node check cpn-[01-40] --arch all
```

This prints a table with a row per node, or per node and arch with `--arch all`. It reads the kernel and initrds at the paths rendered for each node, and renders the iPXE script and every template in the image's `provision_templates`. Files are read on the machine the API runs on. The command exits non-zero when any node fails.

## Ubuntu

In order to use Autoinstall, we will need a template in `/var/lib/grendel/templates/ubuntu-autoinstall.tmpl`. An example file can be found [here](https://github.com/ubccr/grendel/blob/main/internal/render/templates/ubuntu-autoinstall.tmpl), modify as needed.

Then update the grendel image to include a new template: `"user_data": "/var/lib/grendel/templates/ubuntu-autoinstall.tmpl"`

You will need to make sure your cmdline includes the following:
```go
autoinstall cloud-config-url=/dev/null ds=nocloud-net;s={{ $.endpoints.CloudInitURL }}
```

## Rocky

In order to use Kickstart, we will need a template in `/var/lib/grendel/templates/rocky-kickstart.tmpl`. An example file can be found [here](https://github.com/ubccr/grendel/blob/main/internal/render/templates/kickstart.tmpl), modify as needed.

Then update the grendel image to include a new template: `"user_data": "/var/lib/grendel/templates/rocky-kickstart.tmpl"`

The kernel command line will need the following: `inst.ks={{ $.endpoints.KickstartURL }}`.
