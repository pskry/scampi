---
title: ssh
---

Run steps on a remote host via SSH.

```python
target.ssh(
    name = "web",
    host = "app.example.com",
    user = "deploy",
)
```

## Fields

| Field      | Required | Default | Description                    |
| ---------- | :------: | ------- | ------------------------------ |
| `name`     |    ✓     |         | Identifier for deploy blocks   |
| `host`     |    ✓     |         | Hostname or IP address         |
| `user`     |    ✓     |         | SSH user                       |
| `port`     |          | `22`    | SSH port                       |
| `key`      |          |         | Path to private key file       |
| `insecure` |          | `False` | Skip host key verification     |
| `timeout`  |          | `"5s"`  | Connection timeout (Go format) |

## Authentication

SSH targets try authentication methods in order:

1. **Explicit key** — if `key` is set, the private key file is loaded and used.
2. **SSH agent** — if `$SSH_AUTH_SOCK` is set, the agent is queried for keys.

At least one method must succeed. If neither is available, scampi reports an
error with guidance on how to configure authentication.

## Host key verification

By default, scampi verifies host keys against `~/.ssh/known_hosts`. Set
`insecure=True` to skip verification — useful for ephemeral test environments,
but not recommended for production.

## How it works

On connection, the SSH target probes the remote system to detect the OS,
package manager, init system, container runtime, and privilege escalation tool
(sudo/doas). This determines which step capabilities are available.
