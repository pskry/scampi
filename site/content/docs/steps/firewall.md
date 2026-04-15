---
title: firewall
---

Manage firewall rules via UFW or firewalld.

## Fields

| Field       | Type             | Required | Default                  | Description                             |
| ----------- | ---------------- | :------: | ------------------------ | --------------------------------------- |
| `port`      | int              |    ✓     |                          | Port number (1–65535)                   |
| `end_port`  | int?             |          |                          | End of port range (must be > port)      |
| `proto`     | `FirewallProto`  |          | `FirewallProto.tcp`      | Protocol                                |
| `action`    | `FirewallAction` |          | `FirewallAction.allow`   | Rule action                             |
| `desc`      | string?          |          |                          | Human-readable description              |
| `on_change` | list\[Step]      |          |                          | Steps to trigger when this rule changes |

## Actions

`posix.FirewallAction` is an enum:

| Value                   | UFW command         | firewalld command       |
| ----------------------- | ------------------- | ----------------------- |
| `FirewallAction.allow`  | `ufw allow 22/tcp`  | `--add-port=22/tcp`     |
| `FirewallAction.deny`   | `ufw deny 22/tcp`   | rich rule with `drop`   |
| `FirewallAction.reject` | `ufw reject 22/tcp` | rich rule with `reject` |

## Protocols

`posix.FirewallProto` is an enum: `tcp` (default), `udp`.

## How it works

The step manages a single firewall rule per call. It auto-detects the firewall
backend on the target and dispatches to the appropriate tool.

### Backend detection

On every check, scampi probes for a supported backend:

1. `ufw version` — if exit 0, use UFW
2. `firewall-cmd --version` — if exit 0, use firewalld
3. Neither found → error with hint to install one

### UFW

- **Check**: runs `ufw show added` and looks for `ufw <action> <port>`. This
  works even when UFW is inactive — rules are stored, just not enforced. Use the
  `service` step to enable UFW itself.
- **Apply**: runs `ufw <action> <port>`.

### firewalld

- **Check (allow)**: runs `firewall-cmd --query-port=<port>`.
- **Check (deny/reject)**: runs `firewall-cmd --query-rich-rule='...'`.
- **Apply (allow)**: runs `firewall-cmd --permanent --add-port=<port>` then
  `firewall-cmd --reload`.
- **Apply (deny/reject)**: adds a rich rule with `--permanent` then reloads.

The `--permanent` + `--reload` pattern ensures rules persist across reboots.

## Examples

### Allow SSH

```scampi {filename="deploy.scampi"}
posix.firewall {
  port = 22
  desc = "allow SSH"
}
```

### Allow HTTP and HTTPS

```scampi {filename="deploy.scampi"}
posix.firewall { port = 80, desc = "allow HTTP" }
posix.firewall { port = 443, desc = "allow HTTPS" }
```

### UDP rule

```scampi {filename="deploy.scampi"}
posix.firewall {
  port  = 53
  proto = posix.FirewallProto.udp
  desc  = "allow DNS"
}
```

### Port range

```scampi {filename="deploy.scampi"}
posix.firewall {
  port     = 6000
  end_port = 6007
  desc     = "allow X11 forwarding"
}
```

### Deny a port

```scampi {filename="deploy.scampi"}
posix.firewall {
  port   = 3306
  action = posix.FirewallAction.deny
  desc   = "block MySQL from outside"
}
```

### Server hardening pattern

```scampi {filename="harden.scampi"}
posix.pkg {
  packages = ["ufw"]
  state    = posix.PkgState.present
  source   = posix.pkg_system {}
  desc     = "install firewall"
}

posix.firewall { port = 22, desc = "allow SSH" }
posix.firewall { port = 80, desc = "allow HTTP" }
posix.firewall { port = 443, desc = "allow HTTPS" }

posix.service { name = "ufw", state = posix.ServiceState.running, enabled = true }
```
