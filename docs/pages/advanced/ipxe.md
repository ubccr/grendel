# iPXE firmware selection and booting

Grendel ships several embedded iPXE binaries for various architectures. Grendel will attempt to autoselect the correct binary depending on the DHCP Client System Architecture Type (option 93)

| Boot Mode | Client Request | Grendel iPXE Binary | Arch()   | ArchDebian() | ArchMkosi() |
| --------- | -------------- | ------------------- | -------- | ------------ | ----------- |
| BIOS      | Intel x86PC    | undionly.kpxe       | x86_64   | amd64        | x86-64      |
| UEFI      | EFI x86-64     | snponly-x86_64.efi  | x86_64   | amd64        | x86-64      |
| UEFI      | EFI arm64      | snponly-arm64.efi   | aarch64  | arm64        | arm64       |
| UEFI      | EFI i386       | ipxe-i386.efi       | i386     | i386         | x86         |

The arch columns show the values detected for each client request, see [Host architecture](#host-architecture) below.

## Host architecture

The same option 93 value is carried in the node's boot token and used for that provision only. Setting the `arch` field on a node overrides detection.

Boot Tokens created manually via the cli can specify the `--arch <Arch>` flag: `grendel node token cpn-01 boot --arch aarch64`.

`arch` is available to templates as `{{ $.host.Arch }}`, for example to pick a kernel per architecture: `--kernel '/var/lib/grendel/images/rocky10-{{ $.host.Arch }}/vmlinuz'`. Two helpers return the arch in other common naming conventions, and any arch without an alias unchanged:

- `{{ $.host.ArchDebian }}` uses Debian style names
- `{{ $.host.ArchMkosi }}` uses mkosi's names, as found in `Architecture=` and expanded by the `%a` specifier

## Override firmware

If you run into any issues pxe booting, like the kernel download hanging or a grendel error saying `unsupported system client architecture`, you can override the firmware on any node by setting the `firmware` field to one of the following included binaries:

- undionly.kpxe
- ipxe.pxe
- ipxe-i386.efi
- ipxe-x86_64.efi
- snponly-x86_64.efi
- snponly-arm64.efi
