# BMC Compatibility Testing

Grendel talks to BMCs over Redfish, and Redfish is a standard that vendors implement differently.
Grendel's BMC layer was written and validated against Dell iDRACs. On other hardware — or on iDRAC
firmware nobody has tried yet — the failures are quiet ones: a serial number column that comes back
blank, a system event log sorted upside down, a job clear that deletes a URL which does not exist.

`TestCompat` is a suite you point at your own hardware to find out. It exercises every BMC operation
Grendel implements and tells you, capability by capability, whether it works on your machine and if
not which specific assumption fails.

You do not have to run it against live hardware. The recommended path captures a static snapshot of
one BMC, serves that snapshot locally, and runs the suite against the snapshot — so nothing touches
the machine after the initial read.

## 1. Capture a mockup from your BMC

DMTF's mockup creator walks a live Redfish service and writes every resource to disk as
`<uri>/index.json`.

```sh
docker run --rm -v "$PWD/mytree:/output" dmtf/redfish-mockup-creator \
  --rhost 172.18.86.99 --user USERNAME --password PASSWORD \
  --Secure --Dir /output --maxlogentries 50
```

`--Secure` makes it speak HTTPS, which every real BMC requires. `--maxlogentries` caps how much of
each log service it pulls; without it a full system event log adds thousands of resources for no
benefit here.

This is a read-only operation, but it is a thorough one — expect a few minutes and a few thousand
requests against the BMC. Do not run it against a machine in production use.

!!! warning "`-S` means something different to each tool"

    For the **creator**, `-S` is `--Secure` — connect over HTTPS. For the **server** in the next
    step, `-S` is `--short-form` — a differently shaped directory tree. They are unrelated, and
    passing the wrong one is the most common way to get a mockup that will not serve.

!!! warning "The output is large and completely unredacted"

    A stock PowerEdge R760 capture is **91 MB across ~3,350 resources**. Almost all of that is
    `JsonSchemas`, `Schemas` and `Registries`, which Grendel never fetches; the resources it
    actually reads come to about 60 KB.

    Nothing is redacted. Service tags, serial numbers, asset tags, hostnames, IP addresses, MAC
    addresses, iDRAC account names and license keys all ship exactly as captured. **Prune and scrub
    before sharing a tree with anyone**, including on a GitHub issue.

    Note that Dell inlines full log entries into the `LogServices/*/Entries` collection document
    rather than only linking them, so scrubbing the individual entry files is not enough — the
    collection's own `index.json` carries the same data.

## 2. Serve it

```sh
docker run -d --rm --name mockup -p 8000:8000 \
  -v "$PWD/mytree:/mockup:ro" \
  dmtf/redfish-mockup-server -H 0.0.0.0 -D /mockup
```

The creator writes the *full* form, rooted at `<dir>/redfish/v1/...`, which is what `-D` expects —
so do not pass `-S` here. A tree rooted at the contents of `/redfish/v1` instead needs
`-S/--short-form`, and without it the server refuses to start with
`Invalid Mockup Directory--no /redfish directory at top`.

!!! note "Do not use `redfish-interface-emulator`"

    It looks like the right tool and is not. It synthesises its own service root with an empty
    `Links`, so session login cannot work; it loads only eight named subtrees, so `JobService`
    returns 404 whatever your capture contains; and it ignores a mounted mockup entirely unless
    `emulator-config.json` is replaced with one setting `"STATIC": "Enable"` — a key the shipped
    config omits and the code reads inside a bare `except: pass`, so there is no error to tell you
    it was ignored.

## 3. Run the suite

```sh
go test ./internal/redfish -run TestCompat -v \
  -redfish.endpoint=http://localhost:8000 \
  -redfish.basic-auth
```

`-redfish.basic-auth` is required against a mockup server, which authenticates nothing and
advertises no usable session service.

| Flag | Default | Meaning |
|---|---|---|
| `-redfish.endpoint` | *(empty)* | Redfish base URL. **Empty skips the whole suite**, which is why `go test ./...` never runs it |
| `-redfish.user` / `-redfish.password` | `compat` | Credentials. Any non-empty value works against a mockup server |
| `-redfish.basic-auth` | `false` | Log in with basic auth instead of a Redfish session |
| `-redfish.insecure` | `true` | Skip TLS verification, which every BMC needs |
| `-redfish.destructive` | `false` | Also run the checks that write: power reset, job clear, SEL clear |
| `-redfish.bios-change` | *(empty)* | `Name=Value` of a harmless BIOS attribute, such as `NumLock=Off`, to stage and then discard. Needs `-redfish.destructive`. Pick one that would do no harm if the discard failed and the next reboot applied it |

## 4. Read the report

The report is the test output, and it turns on one distinction.

**SKIP means your hardware does not have this.** No job service, no SEL log service, no Dell OEM
block. A real run reports these as a failed node with a message saying what is missing, which is
correct behaviour — a skip is a pass that happens to carry information.

**FAIL means your hardware has it and Grendel gets it wrong.** This is a bug in Grendel, not in your
BMC. Open an issue and attach the `index.json` of the resource named in the failure.

```
--- PASS: TestCompat/status/model
--- FAIL: TestCompat/status/serial_number
        bmc status reports an empty serial_number, so statusOne is
        reading a field this service does not populate
--- SKIP: TestCompat/jobs
        service root declares no JobService
--- FAIL: TestCompat/sel/timestamps
        entry 13 has no timestamp entryTime can read,
        Created="02/23/2026 15:52:27" EventTimestamp=""
```

Lines that are neither are informational, and some are worth reading even when everything passes.
The `power` checks compare what `grendel bmc power` advertises in its help against what your system
actually declares, so a healthy R760 still reports that `ForceOn`, `Suspend`, `Pause`, `Resume`,
`FullPowerCycle` and boot override `Diags` are not available on it.

## What is checked

| Capability | Checks |
|---|---|
| `auth` | Whether Redfish session login and basic auth are each accepted |
| `status` | Every column `grendel bmc status` prints, failing on any that comes back empty |
| `power` | Reset types and boot override targets the system declares, against what the CLI offers; that the manager supports the `GracefulRestart` that `bmc reboot-bmc` sends |
| `jobs` | That the job service exists, that jobs parse, and that the URL Grendel builds to delete a job is where the service actually keeps it |
| `sel` | Which log services are identified as the system event log, on the managers and then the systems; that entry timestamps are in a format Grendel can sort on; that `$top` returns the newest entries; that each one declares a `ClearLog` action |
| `bios` | That the BIOS attributes can be read; whether changes are staged in a separate settings object, which pending values and discard need; whether the BMC serves its attribute registry, and which attributes it leaves undescribed; with `-redfish.bios-change`, that a staged change shows as pending and can be discarded |
| `identity` | That the asset tag and chassis location can be read, and where the location is kept (the iDRAC's ServerTopology attributes, or the chassis `Location`); with `-redfish.destructive`, that a rack written there reads back and is restored |
| `fanout` | That the concurrent runner and error classifier agree with the per-node results above |

!!! note "`bmc sel list --limit` may not show the newest entries"

    Redfish has no way to ask for the newest entries of a log. It defines no `$orderby` and no
    order for the members of a collection. `--limit` is sent to the BMC as `$top`, which returns
    the first entries in whatever order the BMC lists them. Most BMCs list their log newest first,
    but not all of them, and on those `--limit` returns some other set of entries. Grendel sorts
    them newest first either way. `sel/top_is_newest` fails on such a BMC; there, use `--limit 0`
    to read the whole log and sort it. Against a mockup the check only means something if the
    mockup server honours `$top`. A server that ignores it returns the whole log, which Grendel
    sorts correctly, so the check passes.

## Running against live hardware

The suite takes a URL and cannot tell a mockup server from a live iDRAC, so treat
`-redfish.destructive` with care. Without it, nothing that changes machine state is sent — the
power, job clear and SEL clear checks are satisfied by asking the service what it declares support
for. The one exception is `auth`, which creates and deletes a Redfish session, exactly as any
ordinary Grendel run does.

**`-redfish.destructive` will reboot the machine**, clear its job queue and clear its system event
log. Use it on a node you own, or on a mockup.

One caveat if you use it against a mockup: the mockup server answers `ClearLog` with 204 without
removing any entries, so a passing `sel/clear` proves the request was accepted and shaped correctly,
not that the log was cleared. Its `DELETE` handling, on the other hand, does remove the member from
the collection, in memory, for the lifetime of the container — so a destructive run leaves the
mockup changed until you restart it.
