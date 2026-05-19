# ClusterShell

## Configuration

### Grendel can be used as a group source for clush leveraging tags to filter nodes

### See `configs/clush.grendel.conf.sample` for an example clustershell config. The jq should be talored to your compute environment, by default only grendel nodes which start with `cpn` will be returned

## Usage

#### `-g` can be used with a tag name
#### you can also use `-w` with the @ prefix, ie: `-w @<tag>`

```
$ clush -s grendel -b -g rack=f01 -- echo test
---------------
cpn-f01-[04-07,09-12,14-17,22-25,31,33] (18)
---------------
test
```

#### multiple tags can be specified with comma separation
```
$ clush -s grendel -b -g rack=f01,rack=f02 -- echo test
---------------
cpn-f[01-02]-[04-07,09-12,14-17,22-25,31,33] (36)
---------------
test
```

#### tags can be excluded from the list with `-X <tag>`
```
$ clush -s grendel -b -g rack=f01,rack=f02 -X gpu -- echo test
---------------
cpn-f[01-02]-[04-07,09-12,14-17,22-25] (32)
---------------
test
```

#### trying to filter by a tag which doesn't exist will display the grendel API error
```
$ clush -s grendel -b -g randomtag -- echo test
[2026-05-01T11:51:39-04:00] FATAL CLI: API Error: status=500 title=Error detail=failed to filter nodes
Usage: clush [options] command

clush: error: No node to run on.
```

#### excluding an invalid tag will display the error and continue as if `-X <tag>` wasn't passed
```
$ clush -s grendel -b -g rack=f01 -X randomtag -- echo test
[2026-05-01T11:53:23-04:00] FATAL CLI: API Error: status=500 title=Error detail=failed to filter nodes
---------------
cpn-f01-[04-07,09-12,14-17,22-25,31,33] (18)
---------------
test
```
