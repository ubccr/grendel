# Trusted HTTPS and Code signing

Grendel can be setup to provision over HTTPS. This document will describe the
process of setting up a custom certificate authority (CA), rebuilding iPXE to
embed the CA.crt file, running Grendel provisioning server over HTTPS, and code
signing boot images.

!!! note
    These steps require building Grendel from source.

## Install required software and Clone Grendel source code:

```
$ yum install lzma-sdk-devel xz-devel
# or
$ apt install liblzma-dev

```

```
$ go get -u github.com/go-bindata/go-bindata/...
$ git clone --recursive https://github.com/ubccr/grendel
$ cd grendel
```

## Custom CA

Any method for obtaining SSL certificates is supported in Grendel. If you
already have an organizational wide CA or use a commercial CA, all Grendel needs
is the private key and signed certificates. This section will describe setting up
your own custom CA used for signing SSL certificates. Here we use a tool called
[certstrap](https://github.com/square/certstrap). These steps are also outlined
in more detail [here](https://github.com/square/certstrap#certificate-architecture).

### Initialize new certificate authority:

By default all output files will be created in a directory named out:

```
$ certstrap init --common-name "GrendelCA"
```

### Create a new certificate request for the Grendel server

The common name should match the FQDN of your Grendel server:

```
$ certstrap request-cert --common-name grendel.local
```

### Sign and generate the certificate

```
$ certstrap sign grendel.local --CA GrendelCA
```

You should now have 3 files in the out directory that we will configure Grendel
to use to serve HTTPS:

```
out/GrendelCA.crt
out/grendel.local.crt
out/grendel.local.key
```

## Rebuilding iPXE to include custom CA

We now need to rebuild the iPXE firmware to include the custom CA certificate we
created in the previous step. This will allow iPXE to fetch provisioning assets
over trusted HTTPS.

```
$ cd firmware
$ make clean
$ make build-with-ca
$ make bindata
$ cd ..
$ go build .
```

!!! note
    The Makefile assumes your certificate authorities public key is in a file
    named `out/GrendelCA.crt` in the top level directory of the Grendel source.
    If it's located elsewhere adjust the Makefile accordingly.

Now the grendel binary has the new custom built iPXE firmware embedded and is
configured to trust the certificate authority we created in the previous step. 

## Grendel Configuration

Grendel uses a TOML file for configuration. To configure Grendel to use the SSL
certificates created in the previous section, create a grendel.toml file with
the following:

```toml
[provision]
listen = "0.0.0.0:443"
hostname = "grendel.local"
scheme = "https"
cert = "out/grendel.local.crt"
key = "out/grendel.local.key"

[dhcp]
dns_servers = ["192.168.10.254"]
```

!!! note
    DNS resolution is required for HTTPS. You need to ensure the FQDN of your
    grendel server (which is set via the `hostname` in the above TOML) is
    resovlable via dns. You must also set the `dns_servers` in the `dhcp`
    section to the IP address of your DNS server(s). 

For testing with qemu, you can set the DNS server to be the IP address of the
`tap0` device Grendel is listening on (as shown above). Then in the `hosts.json`
file just include a host for Grendel so it will resolve itself like so:

```json
[{
    "name": "tux01",
    "provision": true,
    "boot_image": "flatcar",
    "interfaces": [
        {
            "fqdn": "tux01.local",
            "ip": "192.168.10.12/24",
            "mac": "DE:AD:BE:EF:12:8C"
        }
    ]
},
{
    "name": "grendel",
    "provision": false,
    "interfaces": [
        {
            "fqdn": "grendel.local",
            "ip": "192.168.10.254/24"
        }
    ]
}]
```

### Start services

Now when you run Grendel it should be listening on port 443:

```
sudo ./grendel --verbose -c grendel.toml serve  --hosts hosts.json --images images.json --listen 192.168.10.254
INFO CLI: Using config file: grendel.toml
INFO CLI: Using database path: :memory:
INFO CLI: Successfully loaded 2 hosts
INFO CLI: Successfully loaded 1 boot images
INFO TFTP: Server listening on: 192.168.10.254:69
INFO DHCP: Binding to interface: tap0
INFO PXE: Server listening on: 192.168.10.254:4011
INFO DHCP: Base URL for ipxe: https://grendel.local:443
INFO DNS: Server listening on: 192.168.10.254:53
INFO PROVISION: Listening on https://192.168.10.254:443
INFO PROVISION: ⇨ https server started on 192.168.10.254:443
INFO DHCP: Using DNS servers: [192.168.10.254]
INFO DHCP: Using Domain Search List: []
INFO DHCP: Default lease time: 24h0m0s
INFO DHCP: Default mtu: 1500
INFO DHCP: Using automatic router configuration
INFO DHCP: Netmask: ffffff00
INFO DHCP: Router Octet4: 1
INFO DHCP: Binding to interface: tap0
INFO DHCP: Server listening on: 192.168.10.254:67
INFO API: Listening on unix domain socket: grendel-api.socket
INFO API: ⇨ http server started on grendel-api.socket
```

If you followed the quickstart and are testing with qemu you can test PXE
booting a vm and should now see iPXE downloading the flatcar kernel and initrd
over HTTPS:

```
iPXE 1.0.0+ (3fe68) -- Open Source Network Boot Firmware -- http://ipxe.org
Features: DNS HTTP HTTPS iSCSI TFTP SRP VLAN AoE ELF MBOOT PXE bzImage 
Configuring (net0 de:ad:be:ef:12:8c)... ok
https://grendel.local:443/boot/ipxe... ok
https://grendel.local:443/boot/file/kernel... ok 
```

## Code Signing Boot Images

Grendel (via iPXE) supports code signing, which allows you to verify the
authenticity and integrity of boot images. For more information see the iPXE
docs [here](https://ipxe.org/crypto). Using the custom CA we created in the
previous section we create a new codesigning certificate that's used to
digitally sign and verify boot images.

### Create code signing certificate

```
$ certstrap request-cert --common-name codesigner.local
```

### Sign and generate the certificate

```
$ certstrap sign --codesigning --CA GrendelCA codesigner.local
```

!!! note
    certstrap doesn't currently support codesigning. There's an open PR
    [here](https://github.com/square/certstrap/pull/90). In order to use the
    command above you can build the code from this branch
    [here](https://github.com/ubccr/certstrap/tree/add-code-signing).

You should now have 2 files in the out directory that we will use to sign boot
images:

```
out/codesigner.local.crt
out/codesigner.local.key
```

### Sign Boot Image

We can now sign the boot image files using the code signing cert we created
above. For example, to sign the flatcar kernel and initrd:

```
$ openssl cms -sign -binary -noattr -in flatcar_production_pxe.vmlinuz -signer out/codesinger.local.crt -inkey out/codesigner.local.key -certfile out/GrendelCA.crt -outform DER -out flatcar_production_pxe.vmlinuz.sig

$ openssl cms -sign -binary -noattr -in flatcar_production_pxe_image.cpio.gz -signer codesigner.local.crt -inkey out/codesigner.local.key -certfile out/GrendelCA.crt -outform DER -out flatcar_production_pxe_image.cpio.gz.sig
```

### Create the Boot Image JSON

Add the "Verify" key to the Grendel boot image JSON:

```json
[{
    "name": "flatcar",
    "verify": true,
    "kernel": "flatcar_production_pxe.vmlinuz",
    "initrd": [
        "flatcar_production_pxe_image.cpio.gz"
    ],
    "cmdline": "flatcar.autologin"
}]
```

If you followed the quickstart and are testing with qemu you can test PXE
booting a vm and should now see iPXE verifying the flatcar kernel and initrd
images:

```
iPXE 1.0.0+ (3fe68) -- Open Source Network Boot Firmware -- http://ipxe.org
Features: DNS HTTP HTTPS iSCSI TFTP SRP VLAN AoE ELF MBOOT PXE bzImage
Configuring (net0 de:ad:be:ef:12:8c)... ok
http://192.168.10.254:80/boot/ipxe... ok
http://192.168.10.254:80/boot/file/kernel... ok
http://192.168.10.254:80/boot/file/kernel.sig... ok
http://192.168.10.254:80/boot/file/initrd-0... ok
http://192.168.10.254:80/boot/file/initrd-0.sig... ok
```

!!! tip
    You do not need to enable HTTPS with codesigning. They can be used
    seperately or together depending on your requirements.
