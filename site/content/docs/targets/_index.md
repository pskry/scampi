---
title: Target Reference
weight: 5
---

A target defines where steps execute. Each target type provides a different
transport — local shell, SSH, or HTTP — but they all plug into the same deploy
block mechanism.

```python
target.ssh(name="web", host="app.example.com", user="deploy")

deploy(
    name = "webserver",
    targets = ["web"],
    steps = [ ... ],
)
```

Deploy blocks reference targets by name. A single config can declare multiple
targets of different types and bind them to different deploy blocks.

## Available targets

{{< cards >}}
  {{< card link="local" title="local" subtitle="Run steps on the local machine" >}}
  {{< card link="ssh" title="ssh" subtitle="Run steps on a remote host via SSH" >}}
  {{< card link="rest" title="rest" subtitle="Make HTTP requests against a REST API" >}}
{{< /cards >}}
