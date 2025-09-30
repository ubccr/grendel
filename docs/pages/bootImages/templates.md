# Auto Provisioning

While you can provision a node into a live installer, and finish the install with a keyboard and mouse, repeating this process for thousands of nodes is not ideal. This process is distro specific:

| Distro    | Method       |
|-----------|--------------|
| Debian    | [Preseed](https://wiki.debian.org/DebianInstaller/Preseed)                                            |
| Ubuntu    | [Autoinstall](https://canonical-subiquity.readthedocs-hosted.com/en/latest/intro-to-autoinstall.html) |
| Rocky     | [Anaconda](https://anaconda-installer.readthedocs.io/en/latest/kickstart.html)                        |
| RHEL      | [Anaconda](https://anaconda-installer.readthedocs.io/en/latest/kickstart.html)                        |

## Provision Templates

Kickstart and other provision templates are rendered through Go text templates. You can find some simple example templates [here](https://github.com/ubccr/grendel/tree/main/internal/provision/templates). In these templates, you can call any package functions defined in the text/template package as well as what Grendel adds below:

- [Functions](https://github.com/ubccr/grendel/blob/main/internal/provision/template.go)
- [Variables](https://github.com/ubccr/grendel/blob/main/internal/provision/handler.go#L163)

## Ubuntu

In order to use Autoinstall, we will need a template in `/var/lib/grendel/templates/ubuntu-autoinstall.tmpl`. An example file can be found [here](https://github.com/ubccr/grendel/blob/main/internal/provision/templates/ubuntu-autoinstall.tmpl).

You will need to append the following to the boot images cmdline:
```go
autoinstall cloud-config-url=/dev/null ds=nocloud-net;s={{ $.endpoints.CloudInitURL }}
```