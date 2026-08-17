# Boot Images

PXE booting linux distributions comes in many shapes and sizes, let's run through a simple example of how you can extract the needed files from an ISO.

## Tools

`libcdio` can be used instead of mounting the ISO to extract the required files.

## Ubuntu

We need to extract a few things from the ISO:

- /casper/initrd
- /casper/vmlinuz

 These files can be copied into the `/var/lib/grendel/images` directory, it is recommended to create a new subdirectory for each image type.

 You will also need to place the iso somewhere in the repo directory ex: `/var/lib/grendel/repo/ubuntu`

 ### Creating the Grendel Image:

 You can use the Web UI or CLI to add a new image.
 
 CLI example:

 ```bash
 grendel image add ubuntu \
 --cmdline "console=ttyS0 console=tty0,115200n8 root=/dev/ram0 ramdisk_size=1500000 ip=dhcp url={{ $.endpoints.RepoURL }}/ubuntu/ubuntu-24.04.1-live-server-amd64.iso autoinstall cloud-config-url=/dev/null ds=nocloud-net;s={{ $.endpoints.CloudInitURL }}" \
 --initrd /var/lib/grendel/images/ubuntu/initrd \
 --kernel /var/lib/grendel/images/ubuntu/vmlinuz

 # Provision templates can be added with
 # --provision-template kickstart=/var/lib/grendel/templates/ubuntu-autoinstall.tmpl
 # Or added after the fact with 'grendel image edit <name>'
 ```

 #### Cmdline:

 Base Command line arguments can be found in the grub.cfg file located on the ISO. These may need to be modified depending on the distro for kickstarting to work properly.

 The `cmdline` field is parsed with the same text/template functions and data used in the provision templates.

## Rocky 10

Rocky (like RHEL) boots the Anaconda installer rather than booting the ISO directly. We need to extract the pxeboot kernel and initrd:

- /images/pxeboot/vmlinuz
- /images/pxeboot/initrd.img

These files can be copied into the `/var/lib/grendel/images` directory, it is recommended to create a new subdirectory for each image type.

Unlike the Ubuntu example, Anaconda fetches its installer runtime image (`stage2`) and the packages over HTTP, so you need to serve the **full contents of the ISO** (not just the ISO file) from the repo directory. Mount the ISO and copy/sync its contents into a subdirectory of your `repo_dir`, for example `/var/lib/grendel/repo/rocky/10`:

```bash
mount -o loop,ro Rocky-10.0-x86_64-dvd1.iso /mnt
cp -r /mnt/ /var/lib/grendel/repo/rocky/10/
umount /mnt
```

### Creating the Grendel Image:

You can use the Web UI or CLI to add a new image.

CLI example:

```bash
grendel image add rocky10 \
--cmdline "console=ttyS0 console=tty0,115200n8 ip=dhcp inst.stage2={{ $.endpoints.RepoURL }}/rocky/10 inst.ks={{ $.endpoints.KickstartURL }}" \
--initrd /var/lib/grendel/images/rocky10/initrd.img \
--kernel /var/lib/grendel/images/rocky10/vmlinuz \
--provision-template kickstart=/var/lib/grendel/templates/rocky-kickstart.tmpl
```

#### Cmdline:

- `inst.stage2` points to the directory containing `.treeinfo` / `images/install.img` (the root of the synced ISO). Adjust the path to match your `repo_dir` layout.
- `inst.ks` is rendered to the URL of the kickstart Grendel serves from the `kickstart` provision template (see [Auto Provisioning](templates.md#rocky)).
- `{{ $.endpoints.RepoURL }}` and `{{ $.endpoints.KickstartURL }}` are rendered by Grendel. As with the Ubuntu example, the `cmdline` field is parsed with the same text/template functions and data used in the provision templates.
