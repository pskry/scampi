---
title: sysctl
---

Manage kernel parameters via sysctl with optional persistence.

## Fields

| Field       | Type        | Required | Default | Description                                              |
| ----------- | ----------- | :------: | ------- | -------------------------------------------------------- |
| `key`       | string      |    ✓     |         | Sysctl parameter name (`@std.nonempty`)                  |
| `value`     | string      |    ✓     |         | Desired parameter value (`@std.nonempty`)                |
| `persist`   | bool        |          | `true`  | Write to `/etc/sysctl.d/` for persistence across reboots |
| `desc`      | string?     |          |         | Human-readable description                               |
| `on_change` | list\[Step] |          |         | Steps to trigger when this sysctl changes                |

## How it works

The step runs in two phases:

1. **Set live** — reads the current value with `sysctl -n <key>` and compares
   it to the desired value. If they differ, runs `sysctl -w key=value` to apply
   the change immediately. Reports drift on the key field.

2. **Persist** (when `persist = true`, the default) — checks whether
   `/etc/sysctl.d/99-scampi-<key>.conf` exists with the expected content
   (`key = value`). If not, writes the file. This ensures the setting survives
   reboots. The persist op depends on the set op, so the live value is always
   applied first.

When `persist = false`, only the live value is set and the setting will revert on
reboot. If a scampi-managed drop-in file (matching the `99-scampi-*` naming
convention) exists from a previous run with `persist = true`, it is removed
automatically. scampi only touches its own files — other drop-ins in
`/etc/sysctl.d/` that set the same key are left alone.

## Drop-in file naming

The drop-in filename is derived from the key: dots are replaced with dashes,
and the file is prefixed with `99-scampi-`. Underscores are preserved.

| Key                           | Drop-in file                                               |
| ----------------------------- | ---------------------------------------------------------- |
| `net.ipv4.ip_forward`         | `/etc/sysctl.d/99-scampi-net-ipv4-ip_forward.conf`         |
| `vm.swappiness`               | `/etc/sysctl.d/99-scampi-vm-swappiness.conf`               |
| `net.ipv4.conf.all.rp_filter` | `/etc/sysctl.d/99-scampi-net-ipv4-conf-all-rp_filter.conf` |

The file content is always `key = value` followed by a newline. The `99-`
prefix gives scampi drop-ins high priority — files in `/etc/sysctl.d/` are
applied in lexicographic order, so `99-` overrides distribution defaults that
typically use lower prefixes like `10-` or `50-`.

## Examples

### Enable IP forwarding (persistent)

```scampi {filename="deploy.scampi"}
posix.sysctl {
  key   = "net.ipv4.ip_forward"
  value = "1"
  desc  = "enable IP forwarding"
}
```

Produces:

```ini {filename="/etc/sysctl.d/99-scampi-net-ipv4-ip_forward.conf"}
net.ipv4.ip_forward = 1
```

### Tune TCP keepalive (live only)

```scampi {filename="deploy.scampi"}
posix.sysctl {
  key     = "net.ipv4.tcp_keepalive_time"
  value   = "300"
  persist = false
  desc    = "reduce TCP keepalive interval"
}
```

No drop-in file is written. If one exists from a previous run, it is removed.

### Harden network settings

```scampi {filename="deploy.scampi"}
posix.sysctl { key = "net.ipv4.conf.all.rp_filter", value = "1" }
posix.sysctl { key = "net.ipv4.conf.default.rp_filter", value = "1" }
posix.sysctl { key = "net.ipv4.icmp_echo_ignore_broadcasts", value = "1" }
```
