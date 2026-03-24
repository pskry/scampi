---
title: rest
---

Make HTTP requests against a REST API. Used for configuring applications that
expose a REST interface — reverse proxies, monitoring tools, container
orchestrators, DNS providers.

```python
target.rest(
    name = "npm_api",
    base_url = "http://10.10.2.30:81/api",
    auth = rest.bearer(
        token_endpoint = "/tokens",
        identity = secret("npm.admin.email"),
        secret = secret("npm.admin.password"),
    ),
)
```

## Fields

| Field      | Required | Default             | Description                             |
| ---------- | :------: | ------------------- | --------------------------------------- |
| `name`     |    ✓     |                     | Identifier for deploy blocks            |
| `base_url` |    ✓     |                     | Base URL prepended to all request paths |
| `auth`     |          | `rest.no_auth()`    | Authentication strategy (see below)     |
| `tls`      |          | `rest.tls.secure()` | TLS configuration (see below)           |

## Authentication

Auth strategies are composable — each one wraps the HTTP transport and handles
credentials transparently. New strategies can be added without changing the
target.

### rest.no_auth

No authentication. Requests are sent without any credentials. This is the
default.

```python
auth = rest.no_auth()
```

### rest.basic

HTTP Basic authentication.

```python
auth = rest.basic(user="admin", password=secret("pass"))
```

| Field      | Required | Description |
| ---------- | :------: | ----------- |
| `user`     |    ✓     | Username    |
| `password` |    ✓     | Password    |

### rest.header

Static header authentication. Works for API keys, static bearer tokens, or any
auth that uses a single header.

```python
auth = rest.header(name="X-API-Key", value=secret("grafana.api_key"))
```

| Field   | Required | Description  |
| ------- | :------: | ------------ |
| `name`  |    ✓     | Header name  |
| `value` |    ✓     | Header value |

### rest.bearer

Credential exchange. POSTs identity and secret to a token endpoint, caches the
token, and automatically re-authenticates on 401 responses.

```python
auth = rest.bearer(
    token_endpoint = "/tokens",
    identity = secret("npm.admin.email"),
    secret = secret("npm.admin.password"),
)
```

| Field            | Required | Description                                   |
| ---------------- | :------: | --------------------------------------------- |
| `token_endpoint` |    ✓     | Path to the token endpoint (relative to base) |
| `identity`       |    ✓     | Identity/username for credential exchange     |
| `secret`         |    ✓     | Secret/password for credential exchange       |

The token endpoint must return JSON with a `token` or `access_token` field.

## TLS

TLS strategies are composable, same as auth — one slot, pick the right one.

### rest.tls.secure

Validate certificates against the system CA pool. This is the default.

```python
tls = rest.tls.secure()
```

### rest.tls.insecure

Skip all certificate verification. Use for testing only.

```python
tls = rest.tls.insecure()
```

### rest.tls.ca_cert

Validate against a custom CA certificate. Use for self-signed or internal CAs.

```python
tls = rest.tls.ca_cert(path="./certs/internal-ca.pem")
```

| Field  | Required | Description                        |
| ------ | :------: | ---------------------------------- |
| `path` |    ✓     | Path to PEM-encoded CA certificate |

## How it works

The REST target wraps a standard Go HTTP client. Authentication and TLS are
implemented as composable configuration — each strategy plugs into a single
slot without conflicting fields.

Unlike local and SSH targets, the REST target does not support filesystem,
package, or service operations. Steps that require those capabilities (like
`copy`, `pkg`, or `service`) will fail at plan time with a capability mismatch
error.
