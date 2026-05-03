---
title: pve
---

Run `posix.*` steps inside an LXC container on a Proxmox VE host.
`pve.lxc_target` connects to the PVE host over SSH and proxies every
operation through `pct exec` and `pct push` — no SSH server inside the
container is required, and no network access into the container needs to
be configured.

```scampi
import "std/pve"

let box = pve.lxc_target {
  name = "box"
  host = "10.0.0.5"
  user = "root"
  vmid = 200
}
```

The PVE host itself is reached as a regular `ssh.target`. `pve.lxc_target`
is a separate target type that wraps that SSH connection and projects
POSIX semantics through `pct`. See [Why scampi]({{< relref "../../why" >}})
for the architectural background.

## Fields

| Field      | Type    | Required | Default | Description                        |
| ---------- | ------- | :------: | ------- | ---------------------------------- |
| `name`     | string  |    ✓     |         | Identifier for deploy blocks       |
| `host`     | string  |    ✓     |         | Hostname or IP of the PVE host     |
| `user`     | string  |    ✓     |         | SSH user on the PVE host           |
| `vmid`     | int     |    ✓     |         | LXC container ID (must be ≥ 100)   |
| `port`     | int     |          | `22`    | SSH port on the PVE host           |
| `key`      | string? |          |         | Path to SSH private key file       |
| `insecure` | bool?   |          |         | Skip SSH host key verification     |
| `timeout`  | string  |          | `"5s"`  | SSH connection timeout (Go format) |

## How it works

`pve.lxc_target` satisfies the same POSIX capability contract as
`ssh.target` and `local.target`. Every step in the `posix.*` library —
`posix.pkg`, `posix.copy`, `posix.template`, `posix.service`,
`posix.user`, etc. — runs against it unchanged. The step doesn't know
it's being multiplexed through `pct exec`; the target handles that.

Concretely:

- File reads and writes use `pct push` / `pct pull` from the PVE host
  into the container.
- Command execution uses `pct exec <vmid> -- <command>`.
- File mode, ownership, and stat operations run inside the container's
  filesystem namespace.

This is the same composability pattern that lets `posix.copy` work on a
laptop, a remote SSH host, and an LXC container without writing the step
three times.

## Provisioning + configuring in one file

A common pattern is to create a container with
[`pve.lxc`]({{< relref "../steps/pve/lxc" >}}) and then configure it
with `posix.*` steps in the same scampi run:

```scampi {filename="provision.scampi"}
let pve_host = ssh.target     { name = "pve", host = "10.0.0.5", user = "root" }
let box      = pve.lxc_target { name = "box", host = "10.0.0.5", user = "root", vmid = 200 }

// Create the container on the PVE host.
std.deploy(name = "create-box", targets = [pve_host]) {
  pve.lxc {
    id       = 200
    node     = "pve"
    hostname = "box"
    memory   = "1G"
    networks = [pve.LxcNet { name = "eth0", bridge = "vmbr0", ip = "10.0.0.20/24" }]
  }
}

// Configure the new container with regular posix.* steps.
std.deploy(name = "configure-box", targets = [box]) {
  posix.pkg     { packages = ["nginx"], source = posix.pkg_system {} }
  posix.service { name = "nginx", state = posix.ServiceState.running, enabled = true }
}
```

Two deploy blocks, two targets, one file. The provisioning/configuration
split that other tools enforce as a workflow boundary is, in scampi,
just two targets in the same run.
