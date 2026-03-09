---
title: service
weight: 4
---

Ensure a service is running or stopped, and enabled or disabled at boot.
Works with systemd.

## Fields

| Field     | Type   | Required | Default      | Description |
|-----------|--------|:--------:|--------------|-------------|
| `name`    | string | ✓ |              | Service name |
| `desc`    | string |   |              | Human-readable description |
| `enabled` | bool   |   | `true`       | Whether the service should start at boot |
| `state`   | string |   | `"running"`  | Desired state: `running` or `stopped` |

## How it works

The `service` step produces two independent ops:

1. **Ensure active state** — start or stop the service
2. **Ensure enabled state** — enable or disable at boot

These ops have no dependency on each other and run in parallel.

## Examples

### Start and enable

```python
service(name="nginx", state="running", enabled=True)
```

### Stop and disable

```python
service(name="apache2", state="stopped", enabled=False)
```

### Running but not at boot

```python
service(
    desc = "one-time migration service",
    name = "migrate",
    state = "running",
    enabled = False,
)
```
