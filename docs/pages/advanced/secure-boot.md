# Secure Boot Keys

With UEFI secure boot on, a node's firmware only runs boot loaders and kernels signed by a key in
its `db` database. Machines ship with vendor and Microsoft keys, which cover signed distribution
shims. To boot a kernel or iPXE build you sign yourself, enroll your own certificate in `db`.

Grendel manages secure boot through each node's BMC over Redfish, across a whole nodeset at once.
It keeps no keys itself: the CLI reads the certificate file you give it, and the server enrolls it
on every BMC.

```
grendel bmc secureboot status cpn-a01-[01-10]
```

| Database | Holds |
|---|---|
| `PK` | The platform key, the single key that controls changes to `KEK` |
| `KEK` | Key exchange keys, which control changes to `db` and `dbx` |
| `db` | Certificates and hashes allowed to boot |
| `dbx` | Certificates and hashes revoked, which never boot even if `db` allows them |

Each also has a read-only factory copy, such as `dbDefault`, which `keys --defaults` lists.

## Enrolling your own key

1. Check the current state, and which keys are already enrolled:

    ```
    grendel bmc secureboot status cpn-a01-01
    grendel bmc secureboot keys cpn-a01-01 db
    ```

2. Enroll the certificate of the key you sign with. The file can be PEM, with one or more
   certificates, or a single DER certificate. A node that already has the certificate is left
   alone, so re-running is safe:

    ```
    grendel bmc secureboot add cpn-a01-[01-10] db grendel-db.pem
    ```

3. Turn secure boot on, then reboot the nodes:

    ```
    grendel bmc secureboot enable cpn-a01-[01-10]
    grendel bmc power cycle cpn-a01-[01-10]
    ```

4. After the reboot, `status` shows **Current Boot** as `Enabled`.

Changes take effect at the next boot. `status` shows a staged change under **Pending** on BMCs that
stage it, such as an iDRAC.

## Signing iPXE

With secure boot on, a node's firmware refuses the iPXE build Grendel serves unless it is signed by
a key in `db`. The builds embedded in grendeld are unsigned, so write them out, sign them with your
key, and point grendeld at the signed copies. This runs on the grendeld host, not through the API:

```
grendeld firmware export /var/lib/grendel/firmware
cd /var/lib/grendel/firmware
sbsign --key db.key --cert db.crt --output snponly-x86_64.efi snponly-x86_64.efi
```

`export` writes only the EFI builds, since secure boot does not apply to BIOS boot, and does not
overwrite files already there unless given `--force`. Then set the directory in `grendel.toml` and
restart grendeld:

```toml
[tftp]
firmware_dir = "/var/lib/grendel/firmware"
```

Only the EFI builds are read from the directory. An EFI build missing from it is served from the
embedded copy, so the directory only needs the builds your nodes use, and the BIOS builds
(`undionly.kpxe`, `ipxe.pxe`) are always the embedded ones; one left in the directory is ignored
with a warning. The directory is read when grendeld starts. A file in it that is not an EFI binary
stops grendeld from starting, and an unsigned EFI build is logged as a warning.
`grendeld firmware list` shows where each build will come from, its SHA-256, whether it is signed,
and whether it is still this grendeld's build, which is worth checking before the restart:

```
grendeld firmware list --dir /var/lib/grendel/firmware
```

`Signed` only says the file carries a signature, not whose. Enroll the certificate that signed it in
`db`, as above. Secure boot also applies to the kernel iPXE goes on to boot, so that has to be signed
by a key the node trusts too.

When upgrading Grendel, export and sign again: the directory keeps serving the old iPXE builds until
they are replaced. `list` shows such a build as not `Current`, and grendeld logs a warning for it at
start. The comparison leaves the signature out, so a signed copy of the embedded build is current.

## Removing keys

Each BMC numbers its keys differently, so a key is named by its fingerprint, which is the same on
every node. For a certificate that is its SHA-256, the same as
`openssl x509 -in grendel-db.pem -noout -fingerprint -sha256` prints. For a hash it is the hash
itself. `keys` shows the first 16 digits, and `keys --full` shows the whole fingerprint, which
`delete` needs. Passing the certificate file works too:

```
grendel bmc secureboot delete cpn-a01-[01-10] db grendel-db.pem
```

`--id` removes a key by its id on the BMC instead. Ids differ from node to node, so use it on one
node at a time.

To revoke a binary without removing a certificate, add its hash to `dbx`:

```
grendel bmc secureboot add cpn-a01-[01-10] dbx --hash <sha256>
```

## Resetting

```
grendel bmc secureboot reset cpn-a01-01 --to-default
```

| Flag | Effect |
|---|---|
| `--to-default` | Restores the factory keys in every database, or in the one named. This is the way back for a node that no longer boots because of its keys |
| `--delete-all` | Removes every key, which leaves the node in setup mode, where secure boot enforces nothing until a new `PK` is enrolled |
| `--delete-pk` | Removes only the platform key, which also leaves the node in setup mode |

`disable`, `delete` and `reset` ask for confirmation. Pass `--yes` in a script, since with no
terminal to ask on, they make no change.

## Permissions

Reading the state and keys is open to the `admin`, `user` and `read-only` roles. Every change is
`admin` only, since a wrong key, or a missing one, can leave a node unable to boot. Grant the
**Secure boot** permissions to another role from the web UI's role editor if you need to. Each
change is recorded in the event log, naming the database and the key's fingerprint.

## Vendor notes

- **Dell iDRAC**: the BIOS may refuse key changes unless `SecureBootPolicy` is `Custom`. Check with
  `grendel bmc bios get <nodes> SecureBootPolicy` and change it with `grendel bmc bios set`. Both
  this and enabling secure boot are BIOS changes that apply on the next reboot.
- Firmware in user or deployed mode may refuse changes to `PK` and `KEK` over Redfish. The error the
  BMC gives is shown in the **Message** column.
- Secure boot only applies to nodes that boot in UEFI mode.

Not every BMC exposes secure boot over Redfish. Run the [compatibility suite](bmc-compatibility.md)
against yours to find out what works, including `-redfish.secureboot-cert` to try enrolling and
removing a throwaway certificate.
