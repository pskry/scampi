---
title: container
---

Declarative container management. The runtime is detected automatically on the
target.

## Supported runtimes

| Runtime   | Description                                 |
| --------- | ------------------------------------------- |
| `docker`  | Docker Engine (most common)                 |
| `podman`  | Rootless/daemonless alternative to Docker   |
| `nerdctl` | CLI for containerd (ships with k3s toolbox) |
| `finch`   | AWS container CLI (containerd-backed)       |

Detection order is top to bottom — the first one found wins. All four
implement the same Docker-compatible CLI surface, so all container steps
work identically regardless of which runtime is present.

If the runtime works without root privileges (e.g. rootless setup or
appropriate group membership), commands run unprivileged. Otherwise,
escalation (sudo/doas) is used.

## Steps

{{< cards >}}
  {{< card link="instance" title="instance" subtitle="Manage container lifecycle: running, stopped, or absent" >}}
{{< /cards >}}
