# iPXE firmware selection and booting

Grendel ships several embedded iPXE binaries for various architectures. Grendel will attempt to autoselect the correct binary depending on the DHCP Client System Architecture Type (option 93)

| Boot Mode | Client Request | Grendel iPXE Binary | `{{ $.host.Arch }}` | `{{ $.host.ArchDebian }}` | `{{ $.host.ArchMkosi }}` |
| --------- | -------------- | ------------------- | ------------------- | ------------------------- | ------------------------ |
| BIOS      | Intel x86PC    | undionly.kpxe       | x86_64              | amd64                     | x86-64                   |
| UEFI      | EFI x86-64     | snponly-x86_64.efi  | x86_64              | amd64                     | x86-64                   |
| UEFI      | EFI arm64      | snponly-arm64.efi   | aarch64             | arm64                     | arm64                    |
| UEFI      | EFI i386       | ipxe-i386.efi       | i386                | i386                      | x86                      |

The arch columns show the values detected for each client request, see [Host architecture](#host-architecture) below.

## Host architecture

The same option 93 value is used to detect a node's architecture (`x86_64`, `aarch64` or `i386`) while it boots. The detected value is carried in the node's boot token and used for that provision only, it's never saved to the node. Setting the `arch` field on a node overrides detection.

Detection needs Grendel's DHCP to hand iPXE its boot script. Tokens created any other way carry no detected arch, so requests made with them only see a manually set `arch`. To test templates as if a node had booted with a given arch, pass it when creating a token: `grendel node token cpn-01 boot --arch aarch64`.

`arch` isn't used by Grendel itself, it's available to templates as `{{ $.host.Arch }}`, for example to pick a kernel per architecture: `--kernel '/var/lib/grendel/images/rocky10-{{ $.host.Arch }}/vmlinuz'`. Two helpers return the arch in other common naming conventions, and any arch without an alias unchanged:

- `{{ $.host.ArchDebian }}` uses Debian/Go style names: `amd64` for `x86_64` and `arm64` for `aarch64`.
- `{{ $.host.ArchMkosi }}` uses mkosi's names, as found in `Architecture=` and expanded by the `%a` specifier: `x86-64` for `x86_64`, `arm64` for `aarch64`, `x86` for `i386` and `ppc64-le` for `ppc64le`. This is useful for fetching mkosi built images whose names include `%a`, for example `{{ $.endpoints.RepoURL }}/sysext/frosty-sysext-gpu-nvidia-{{ $.host.ArchMkosi }}.raw`.

## Override firmware

If you run into any issues pxe booting, like the kernel download hanging or a grendel error saying `unsupported system client architecture`, you can override the firmware on any node by setting the `firmware` field to one of the following included binaries:

- undionly.kpxe
- ipxe.pxe
- ipxe-i386.efi
- ipxe-x86_64.efi
- snponly-x86_64.efi
- snponly-arm64.efi
