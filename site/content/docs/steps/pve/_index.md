---
title: pve
---

Steps for managing Proxmox VE hosts and containers. All `pve.*` steps run
against an `ssh.target` pointed at a PVE node — they shell into it and
call `pct`, `pvesh`, or write to `/etc/pve/*`.

The companion target type
[`pve.lxc_target`]({{< relref "../../targets/pve" >}}) lets you run any
`posix.*` step inside a container created by `pve.lxc`, without needing
an SSH server in the container.

## Steps

{{< cards >}}
  {{< card link="lxc" title="lxc" subtitle="LXC container lifecycle on a PVE host" >}}
  {{< card link="datacenter" title="datacenter" subtitle="Declarative /etc/pve/datacenter.cfg" >}}
{{< /cards >}}
