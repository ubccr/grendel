# Dynamic DHCP Router Option 3 

Grendel's built in DHCP server can be configured to dynamically set DHCP router
option 3. This can be useful in situations where network configurations are
consistent across many racks of nodes.

For example, suppose Grendel is being used to provision 3 racks of compute
nodes. Each rack is setup on it's own `/24` subnet:

- rack1 = 10.64.25.0/24
- rack2 = 10.64.26.0/24
- rack3 = 10.64.27.0/24

Additionally, each network is setup in a consistent way where the routers are
the same IP address within the network. Suppose the router IP's for each network
are:

- rack1 = 10.64.25.254
- rack2 = 10.64.26.254
- rack3 = 10.64.27.254


Out of the box, Grendel serves static DHCP leases for any hosts defined in the
database. Using the IP address of the compute node, we can dynamically set the
router by setting the last octet to 254. Unlike other DHCP servers there's no
need to define subnet definitions or any other configurations. 

To configure Grendel to dynamicaly set router option 3 when serving static
leases add the following to `grendel.toml`:

```toml
[dhcp]
router_octet4 = 254
netmask = 24
```

Where `router_octet4` is the last octet of the router IP address and `netmask`
is the netmask.

!!! note
    Currently only `/24` subnets are supported


With the above configuration, when Grendel serves a static DHCP lease for
compute node with IP address `10.64.26.12` it will set router option 3 to
`10.64.26.254` in the DHCP respose. 
