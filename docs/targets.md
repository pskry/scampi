# Target Definitions & Configuration Model

This document captures the agreed-upon design for handling **targets**, **steps**, and **credentials** in CUE-based configs.

It is intentionally concise and opinionated.

---

## Core Principle

> **Targets describe *where* and *how* execution happens. Steps describe *what* the desired state is.**

These concerns must remain separate to preserve portability, reuse, and security.

---

## Conceptual Layers

```
Steps        → desired state (what)
Targets      → execution context (where/how)
Bindings     → intent to apply (what to where)
Credentials  → runtime identity (secrets)
```

Each layer has a different lifecycle and ownership model.

---

## Steps (WHAT)

Steps are:

* Portable
* Reusable
* Free of infrastructure concerns

They **must not** reference:

* Hosts
* Environments
* Credentials

Example:

```cue
steps: web: {
  type: "service"
  image: "nginx"
  ports: [80]
}
```

---

## Targets (WHERE / HOW)

Targets describe execution surfaces such as:

* Hosts
* Clusters
* Environments

They may include:

* Connection metadata
* Roles / labels
* Execution type (ssh, local, etc.)

They **must not** include secrets.

Example:

```cue
targets: prod_us_east: {
  type: "ssh"
  host: "10.0.0.10"
  user: "deploy"
  roles: ["web", "db"]
}
```

Targets are expected to be **shared and reused** across configs and projects.

---

## Bindings (INTENT)

Bindings connect steps to targets.

This is where environment-specific intent lives.

Example (explicit):

```cue
deploy: {
  steps: ["web"]
  targets: ["prod_us_east"]
}
```

Example (selector-based):

```cue
deploy: {
  steps: ["web"]
  targetSelector: {
    roles: ["web"]
  }
}
```

Bindings are expected to change frequently; steps and targets are not.

---

## Target Reuse

Targets should live in centralized, importable files.

Recommended layout:

```
targets/
  aws.cue
  baremetal.cue
  local.cue

env/
  prod.cue
  staging.cue
```

Example:

```cue
// env/prod.cue
import "targets/aws"

targets: aws.targets
```

---

## Credentials (SECRETS)

**Credentials never live in CUE configs.**

CUE may reference credentials symbolically, but resolution happens at runtime.

Example:

```cue
targets: prod: {
  type: "ssh"
  host: "10.0.0.10"
  user: "deploy"

  auth: {
    method: "ssh-agent"
    keyRef: "prod-deploy"
  }
}
```

Resolution strategies include:

* SSH agent
* OS keychain
* Vault / secret manager
* Environment variables

The engine must never serialize resolved secrets back into CUE.

---

## Single-File Configs

Inlining steps, targets, and bindings in one file is allowed for:

* Demos
* Tests
* Local bootstrap

It is **not recommended** for reusable or production configs.

Rule of thumb:

> If you expect reuse, do not inline targets.

---

## Summary Rules

* Steps define **what**
* Targets define **where/how**
* Bindings define **intent**
* Credentials are **runtime-only**

Keeping these layers separate is critical for scalability, security, and clarity.
