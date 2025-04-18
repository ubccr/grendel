# iPXE firmware selection and booting

Grendel ships several embedded iPXE binaries for various architectures. Grendel will attempt to autoselect the correct binary depending on the DHCP Client System Architecture Type (option 93)

| Boot Mode | Architecture | Client Request | Grendel iPXE Binary |
| --------- | ------------ | -------------- | ------------------- |
| BIOS      | amd64        | Intel x86PC    | undionly.kpxe       |
| UEFI      | amd64        | EFI x86-64     | snponly-x86_64.efi  |
| UEFI      | arm64        | EFI arm64      | snponly-arm64.efi   |
| UEFI      | i386         | EFI i386       | ipxe-i386.efi       |

## Override firmware

If you run into any issues pxe booting, like the kernel download hanging or a grendel error saying `unsupported system client architecture`, you can override the firmware on any node by setting the `firmware` field to one of the following included binaries:

- undionly.kpxe
- ipxe.pxe
- ipxe-i386.efi
- ipxe-x86_64.efi
- snponly-x86_64.efi
- snponly-arm64.efi
