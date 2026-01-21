# Units vs Steps Model

This document captures the clarified execution model that resolves past issues with naming, ordering, and reuse.

It reflects the current direction intentionally optimized for a **solo developer**, prioritizing conceptual correctness over backward compatibility.

---

## The Core Problem We Hit Before

Historically, the system modeled *everything executable* as a **Unit**.

This caused two major issues:

1. **Forced naming**
   Small, repetitive operations (e.g. copying files) required artificial names like `copy1`, `copy23`.

2. **Accidental ordering semantics**
   Flat lists of units implicitly encoded execution order, even when the intent was declarative convergence.

These were not incidental bugs — they were symptoms of a flawed abstraction boundary.

---

## The Key Insight

> **Not everything executable deserves identity.**

There are two fundamentally different concepts that were previously conflated:

| Concept | Has identity? | Needs ordering? | Example                |
| ------- | ------------- | --------------- | ---------------------- |
| Unit    | Yes           | No              | nginx, postgres, users |
| Step    | No            | Yes             | copy file, run command |

Trying to model *steps* as *units* is what caused the earlier pain.

---

## What a Unit Is (Now)

A **Unit** is:

* Conceptually meaningful
* Stable over time
* Referencable by name
* A convergence boundary

Units represent *what it means for something to be in the desired state*.

### Units are named and sparse

```cue
units: nginx: {
  ...
}
```

Units should **not** be tiny procedural actions.

If naming a unit feels awkward, that is a strong signal it should not be a unit.

---

## What a Step Is

A **Step** is:

* Procedural
* Anonymous
* Order-dependent
* Local to a unit

Steps represent *how a unit converges today*.

### Steps are ordered lists

```cue
steps: [
  { copy: { src: "nginx.conf", dest: "/etc/nginx/nginx.conf" } },
  { run:  { cmd: "nginx -t" } },
  { service: { name: "nginx", state: "restarted" } },
]
```

Steps do **not** have names and do **not** carry identity.

---

## Reframing Builtins (Copy Example)

Previously:

* `copy` was modeled as a **Unit**
* Required a `name`
* Lived at the top level

### New model

* `copy` is a **builtin Step kind**
* It is used *inside* units
* It has no identity of its own

```cue
{ copy: {
  src:  string
  dest: string
  perm?: string
  owner?: string
  group?: string
}}
```

This preserves all execution logic while eliminating artificial naming.

---

## Ordering Semantics

* **Steps inside a unit are ordered** (lists)
* **Units themselves are unordered** (maps)

This keeps ordering scoped to where it is meaningful and avoids encoding global execution order accidentally.

Future DAG-based planning can be introduced without changing this model.

---

## Design Rules (Hard Guardrails)

* If something forces you to invent names, it is not a unit
* Units describe *what*, steps describe *how*
* Identity belongs only at convergence boundaries
* Ordering belongs only to procedural sequences

---

## Summary

* Units are named, semantic, and stable
* Steps are anonymous, ordered, and procedural
* `copy` is a step, not a unit
* This model directly addresses the historical pain points

This separation is foundational and should be preserved even as the system grows.
