---
title: datacenter
---

Render `/etc/pve/datacenter.cfg` from declared fields. Manages PVE
datacenter-level configuration: console viewer, default keyboard, MAC
prefix, datacenter tags with colors, and so on.

```scampi {filename="dc.scampi"}
import "std/pve"

std.deploy(name = "pve-dc", targets = [pve_host]) {
  pve.datacenter {
    console     = pve.Console.html5
    description = "homelab"
    tags        = [
      pve.Tag { name = "prod", fg = "#ffffff", bg = "#1a73e8" },
      pve.Tag { name = "dev",  fg = "#000000", bg = "#fbbc04" },
    ]
  }
}
```

## Fields

| Field         | Type          | Default           | Description                                                     |
| ------------- | ------------- | ----------------- | --------------------------------------------------------------- |
| `console`     | `Console`     | `Console.xtermjs` | Default console viewer: `vv`, `html5`, or `xtermjs`             |
| `keyboard`    | string?       |                   | Default keyboard layout                                         |
| `language`    | string?       |                   | Web UI language                                                 |
| `mac_prefix`  | string?       |                   | OUI prefix for generated MACs (regex `[a-f0-9]{2}:[a-f0-9]{2}`) |
| `max_workers` | int?          |                   | Max concurrent workers per node (must be ≥ 1)                   |
| `email_from`  | string?       |                   | Sender address for PVE notifications                            |
| `http_proxy`  | string?       |                   | Outbound HTTP proxy                                             |
| `description` | string?       |                   | Datacenter description                                          |
| `tags`        | `list\[Tag]`  | `[]`              | Datacenter tags with foreground/background colors               |
| `backup`      | bool          | `true`            | Back up the existing config before writing                      |
| `desc`        | string?       |                   | Human-readable description                                      |
| `on_change`   | `list\[Step]` | `[]`              | Steps to trigger when this config changes                       |

## `Tag` — datacenter tag

```scampi
pve.Tag {
  name = "prod"
  fg   = "#ffffff"   // text colour (hex)
  bg   = "#1a73e8"   // background colour (hex)
}
```

## How it works

The step renders a Go-templated `datacenter.cfg` and writes it to
`/etc/pve/datacenter.cfg` with mode `0640`, owner `root`, group
`www-data`. By default, the existing config is backed up before writing
(`backup = true`).

Only set fields are emitted. Unset fields are left out of the file
entirely — PVE uses its built-in defaults for those.
