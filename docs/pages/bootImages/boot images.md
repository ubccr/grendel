# Boot Images

PXE booting linux distributions comes in many shapes and sizes, let's run through a simple example of how you can extract the needed files from an ISO.

## Tools

`libcdio` can be used instead of mounting the ISO to extract the required files.

## Ubuntu

We need to extract a few things from the ISO:

- /boot/initrd.img
- /boot/linux

 These files can be copied into the `/var/lib/grendel/images` directory, it is recommended to create a new subdirectory for each image type.

 You will also need to place the iso somewhere in the repo directory ex: `/var/lib/grendel/repo/ubuntu`

 ### Creating the Grendel Image:

 You can use the Web UI or CLI to add a new image.
 
 CLI example:

 ```bash
 grendel image add ubuntu \
 --cmdline "console=ttyS0 console=tty0,115200n8 root=/dev/ram0 ramdisk_size=1500000 ip=dhcp url={{ $.endpoints.RepoURL }}/ubuntu/ubuntu-24.04.1-live-server-amd64.iso autoinstall cloud-config-url=/dev/null ds=nocloud-net;s={{ $.endpoints.CloudInitURL }}" \
 --initrd /var/lib/grendel/images/ubuntu/initrd.img \
 --kernel /var/lib/grendel/images/ubuntu/linux

 # Provision templates can be added with
 # --provision-template kickstart=/var/lib/grendel/templates/ubuntu-autoinstall.tmpl
 # Or added after the fact with 'grendel image edit <name>'
 ```

 #### Cmdline:

 Base Command line arguments can be found in the grub.cfg file located on the ISO. These may need to be modified depending on the distro for kickstarting to work properly.

 The `cmdline` field is parsed with the same text/template functions and data used in the provision templates.