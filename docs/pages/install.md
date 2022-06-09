# Installation

This section describes the various methods of installing Grendel.

## Install the pre-compiled binary

Download from the [Grendel releases page][https://github.com/ubccr/grendel/releases].

### tar.gz archive

```
$ tar xvzf grendel-VERSION-linux-x86_64.tar.gz 
```

### deb, rpm packages

```
$ sudo dpkg -i grendel_VERSION_amd64.deb
```

```
$ sudo rpm -ivh grendel-VERSION-amd64.rpm
```

## Configure grendel


!!! tip
    Stock ubuntu runs a local stub dns resolver bound to port 53. If you want
    to run Grendel's built in dns server you will have to free up this port.

### How to free up port 53 used by systemd-resolved

1. Check if port 53 is in use on your system

```
$ sudo lsof -i :53
systemd-r 1261 systemd-resolve   13u  IPv4  28862      0t0  UDP localhost:domain
systemd-r 1261 systemd-resolve   14u  IPv4  28863      0t0  TCP localhost:domain (LISTEN)
```

1. Edit /etc/systemd/resolved.conf: 

```
# Set this to the DNS server you want to use
DNS=1.1.1.1

# Set this to no
DNSStubListener=no
```

1. Update symlink:

```
$ sudo ln -sf /run/systemd/resolve/resolv.conf /etc/resolv.conf
```

1. reboot
