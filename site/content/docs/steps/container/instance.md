---
title: instance
---

Manage container lifecycle: running, stopped, or absent. See the
[container module overview](../) for supported runtimes.

> [!TIP]
> Named volumes and networks are planned for future releases.

## Fields

| Field         | Type        | Required | Default            | Description                           |
| ------------- | ----------- | :------: | ------------------ | ------------------------------------- |
| `name`        | string      |    ✓     |                    | Container name                        |
| `image`       | string      |  ✓[^1]   |                    | Container image (tag or digest)       |
| `desc`        | string      |          |                    | Human-readable description            |
| `state`       | string      |          | `"running"`        | Desired [state](#states)              |
| `restart`     | string      |          | `"unless-stopped"` | [Restart policy](#restart-policies)   |
| `ports`       | list        |          |                    | [Port mappings](#port-mappings)       |
| `env`         | dict        |          |                    | Environment variables                 |
| `mounts`      | list        |          |                    | Bind mounts (`"host:container[:ro]"`) |
| `args`        | list        |          |                    | Arguments for container entrypoint    |
| `labels`      | dict        |          |                    | Container labels                      |
| `healthcheck` | healthcheck |          |                    | [Healthcheck config](#healthchecks)   |

[^1]: Required when state is `running` or `stopped`, optional when `absent`.

## States

| State     | Behavior                                    |
| --------- | ------------------------------------------- |
| `running` | Create and start. Recreate on config drift. |
| `stopped` | Create but don't start. Recreate on drift.  |
| `absent`  | Stop and remove.                            |

## Restart policies

Controls what happens when the container process exits or the host reboots.
Manually stopping a container always works regardless of restart policy —
the policy only governs automatic restarts.

| Policy           | On container exit | On host reboot | On manual stop                       |
| ---------------- | ----------------- | -------------- | ------------------------------------ |
| `always`         | Restart           | Restart        | Stays stopped until next host reboot |
| `unless-stopped` | Restart           | Restart        | Stays stopped permanently            |
| `on-failure`     | Restart           | Do not restart | Stays stopped permanently            |
| `no`             | Do not restart    | Do not restart | Stays stopped permanently            |

The difference between `always` and `unless-stopped`: both restart on
container exit and host reboot, but if you manually stop an `always`
container, it comes back after the next reboot. An `unless-stopped`
container stays down once you stop it.

### Why `unless-stopped` is the default

When someone manually stops a container, they have a reason — debugging,
an incident, a migration. The runtime should respect that decision and
leave the container alone. If the service needs to come back, the operator
runs scampi, which sees the container is stopped while the declared state
is `running`, and starts it — with a visible change in the output.

This keeps the responsibility clear: the restart policy handles crash
recovery (process exits unexpectedly → restart automatically). Scampi
handles convergence (declared state says running → make it so). Manual
interventions are respected until scampi explicitly overrides them.

With `always`, a manual stop is silently undone on the next reboot. That's
surprising — you stopped something and it came back on its own, without
anyone running scampi. For a convergence tool where explicit changes are a
core principle, that's the wrong default.

## Port mappings

Ports are specified as strings in the format `"hostPort:containerPort"`:

```python
ports = ["8080:80", "9090:9090"]
```

Extended formats:

| Format                            | Example                 | Description           |
| --------------------------------- | ----------------------- | --------------------- |
| `hostPort:containerPort`          | `"8080:80"`             | Bind to all addresses |
| `ip:hostPort:containerPort`       | `"127.0.0.1:8080:80"`   | Bind to specific IP   |
| `hostPort:containerPort/proto`    | `"5000:5000/udp"`       | UDP instead of TCP    |
| `ip:hostPort:containerPort/proto` | `"127.0.0.1:53:53/udp"` | IP + UDP              |

TCP is the default protocol. All four fields are preserved across check,
drift detection, and recreate.

## Healthchecks

> [!IMPORTANT]
> Health check commands are currently executed as shell commands
> (`/bin/sh -c`) inside the container. The image must have `/bin/sh`
> available — distroless and scratch-based images cannot use healthchecks
> yet. Exec-form healthchecks (without a shell) are planned.

Use `container.healthcheck.cmd()` to define a health probe:

```python
container.instance(
    name = "app",
    image = "myapp:latest",
    healthcheck = container.healthcheck.cmd(
        cmd = "curl -f http://localhost:8080/health",
        interval = "10s",
        timeout = "5s",
        retries = 5,
    ),
)
```

| Field          | Type   | Default | Description                         |
| -------------- | ------ | ------- | ----------------------------------- |
| `cmd`          | string | —       | Health check command (required)     |
| `interval`     | string | `"30s"` | Time between checks                 |
| `timeout`      | string | `"30s"` | Maximum time for one check          |
| `retries`      | int    | `3`     | Failures before reporting unhealthy |
| `start_period` | string | `"0s"`  | Grace period before checks start    |

Duration fields accept Go duration strings: `"300ms"`, `"1.5s"`, `"2m30s"`,
`"1h"`. Valid units are `ns`, `us`/`µs`, `ms`, `s`, `m`, `h`.

Defaults are set explicitly by scampi, not inherited from the container
runtime, so behavior is identical across runtimes.

When a healthcheck is defined and `state = "running"`, scampi waits for
the container to report healthy after starting it. If the container
becomes unhealthy, the apply fails with a diagnostic.

## How it works

The step produces a single op that handles the full lifecycle:

1. **Check**: inspect the container. Compare image, restart policy, ports,
   environment variables, bind mounts, args, labels, and healthcheck
   against the declared config. Any drift → unsatisfied.
2. **Execute**: depending on the desired state and current state:
   - **Create**: create with the declared config, then start
   - **Recreate**: stop → remove → create → start
   - **Remove**: stop → remove

Containers are **immutable** — any config drift triggers a full recreate cycle.
There are no in-place updates.

## Examples

### Run a container

```python
container.instance(
    name = "prometheus",
    image = "prom/prometheus:v3.2.0",
    ports = ["9090:9090"],
)
```

### Pin an image version

```python
container.instance(
    name = "grafana",
    image = "grafana/grafana:11.5.2",
    ports = ["3000:3000"],
    restart = "unless-stopped",
)
```

### Pass environment variables

```python
container.instance(
    name = "app",
    image = "myapp:latest",
    env = {"DB_HOST": "db.local", "DB_PORT": "5432"},
)
```

Only declared variables are checked for drift — extra variables set by the
base image are ignored.

### Bind mount host directories

```python
dir(path = "/opt/prometheus/data")

container.instance(
    name = "prometheus",
    image = "prom/prometheus:v3.2.0",
    mounts = ["/opt/prometheus/data:/prometheus"],
)
```

Host directories are **not** created by the container step — use `dir()`
before it. The engine automatically orders the `dir` step before the
container step via resource dependencies.

Append `:ro` to make the mount read-only:

```python
mounts = ["/opt/config:/etc/app:ro"],
```

### Pass arguments to the entrypoint

```python
container.instance(
    name = "prometheus",
    image = "prom/prometheus:v3.2.0",
    args = [
        "--config.file=/etc/prometheus/prometheus.yml",
        "--storage.tsdb.retention.time=30d",
    ],
)
```

Arguments are passed to the container's entrypoint. If `args` is not
declared, the image's default command is left untouched and not checked
for drift.

### Add labels

```python
container.instance(
    name = "app",
    image = "myapp:latest",
    labels = {"app": "myapp", "env": "production"},
)
```

Only declared labels are checked for drift — labels added by the base
image are ignored.

### Remove a container

```python
container.instance(
    name = "old-service",
    state = "absent",
)
```
