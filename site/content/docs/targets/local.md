---
title: local
---

Run steps on the machine where scampi is invoked.

```python
target.local(name="my-machine")
```

## Fields

| Field  | Required | Description                            |
| ------ | :------: | -------------------------------------- |
| `name` |    ✓     | Identifier referenced by deploy blocks |

## How it works

The local target executes commands directly on the host. It detects the OS,
package manager, init system, container runtime, and privilege escalation tool
(sudo/doas) automatically.
