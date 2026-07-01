# Web UI & API

Grendel exposes a HTTP API that powers both the `grendel` CLI and the built-in Web UI. The same routes can be served on two different kinds of listener at the same time:

- **UNIX socket** for local access (UNIX permissions only)
- **TCP socket** for remote access (authentication required, HTTPS recommended)

## Listeners

### UNIX socket

```toml
[api]
socket_path = "grendel-api.socket"
```

When `socket_path` is set, Grendel binds the API to a UNIX socket. The socket file is created with `0770` permissions, access is controlled by normal filesystem ownership/group membership.

### TCP socket

```toml
[api]
listen = "127.0.0.1:8080"

# Setting both key and cert enables HTTPS
key  = "/etc/grendel/api/hostname.key"
cert = "/etc/grendel/api/hostname.crt"
```

`listen` binds the API to a TCP address. If both `key` and `cert` are set the listener uses HTTPS.

### Binding both at once

Both listeners can run simultaneously. A common setup is a UNIX socket for local CLI/admin use and a TCP listener (HTTPS) for the Web UI and remote clients.

UNIX requests will be logged with the local username prefixed by unix ie: `unix:ubuntu`

## Authentication

Requests arriving over the UNIX socket are treated as trusted and skip authentication entirely — access is gated by who can open the socket file. There is no role based access control on these requests.

The TCP listener enforces token based authentication with Role Based Access Control (RBAC). Each API route must be explicitly added to the users role for them to have access.

The first user account created will be enabled with the admin role. Any other accounts will be disabled with the user role. The admin user can update other users roles and status via the CLI or Web UI.

Tokens are JSON Web Tokens (JWTs) signed with `api.secret`.

>[!NOTE]
    If `api.secret` is not set, a new signing key is used on each start and all
    previously issued tokens become invalid.

### RBAC

Grendel provides three built in roles: `admin`, `user`, and `read-only`. The permissions on these roles will be updated accrodingly when new versions add authorized API endpoints. New roles can be added with restrictions on any API endpoint you want.

If a role has access to an endpoint, there is no restriction on which nodes they can interact with using that endpoint.

## Issuing tokens

Clients authenticate with a token rather than a username/password. Create one from the Web UI, or CLI:


- `username` - valid user
- `role` - the role the token is granted, e.g. `admin` or `user`
- `expire` - `infinite`, or a duration string parsed by Go's `time.ParseDuration` such as `8h` or `30m`

>[!NOTE]
  You may only request a role whose permissions are less than or equal to your own, and only users with an admin role may issue tokens for a username other than their own.

## Using a token with the CLI

The CLI reads its endpoint and credentials from the `[client]` section.

If connecting over a TCP listener, use the API URL and supply the token as `api_key`:

```toml
[client]
api_endpoint = "https://hostname:8080"
api_key      = "eyJ..."
```
