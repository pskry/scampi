---
title: resource_set
---

Declarative set reconciliation. Fetches the full remote collection with a
single query, matches each remote item against a declared set by key, and
fires per-item actions for missing, drifted, and orphaned entries. Config
becomes the truth — remove a line, next apply cleans up the remote side.

```scampi
rest.resource_set {
  desc  = "fixed IP clients"
  query = rest.request {
    method = "GET"
    path   = "/api/s/default/rest/user"
    check  = rest.jq { expr = ".data[] | select(.use_fixedip == true)" }
  }
  key = rest.jq { expr = ".mac" }
  items = [
    {"mac": "aa:bb:cc:dd:ee:01", "name": "server1", "fixed_ip": "10.0.0.10", "use_fixedip": true},
    {"mac": "aa:bb:cc:dd:ee:02", "name": "server2", "fixed_ip": "10.0.0.11", "use_fixedip": true},
  ]
  missing = rest.request { method = "POST", path = "/api/s/default/rest/user" }
  found   = rest.request { method = "PUT",  path = "/api/s/default/rest/user/{id}" }
  orphan  = rest.request { method = "PUT",  path = "/api/s/default/rest/user/{id}" }
  bindings     = {"id": rest.jq { expr = "._id" }}
  orphan_state = {"use_fixedip": false, "fixed_ip": ""}
}
```

## Fields

| Field          | Type                      | Required | Description                                             |
| -------------- | ------------------------- | :------: | ------------------------------------------------------- |
| `query`        | `rest.request`            |    ✓     | Request to fetch the full remote set                    |
| `key`          | `rest.Check`              |    ✓     | jq expression to extract the match key from each item   |
| `items`        | list\[map\[string, any]]  |          | Desired set of items (empty = everything is an orphan)  |
| `missing`      | `rest.request?`           |          | Request for items in declared set but not remote        |
| `found`        | `rest.request?`           |          | Request for items in both sets with drift               |
| `orphan`       | `rest.request?`           |          | Request for items in remote but not declared            |
| `bindings`     | map\[string, rest.Check]? |          | Per-item jq bindings for path interpolation             |
| `orphan_state` | map\[string, any]?        |          | State to send as body for orphan items                  |
| `desc`         | string?                   |          | Human-readable description                              |
| `on_change`    | list\[Step]               |          | Steps to trigger when any item changes                  |

At least one of `missing`, `found`, or `orphan` is required.

## Query and key

The `query` fetches the full remote set. Its `rest.jq` check filters the
response down to the items to reconcile — typically something like
`.data[]` or `.[] | select(.active == true)`.

The `key` is a `rest.jq` expression that extracts a unique identifier from
each item. It runs against both remote items (from the query) and declared
items (from `items`), producing a string key for matching.

```scampi
query = rest.request {
  method = "GET"
  path   = "/api/users"
  check  = rest.jq { expr = ".data[]" }
}
key = rest.jq { expr = ".mac" }
```

The query fires once, not per-item.

## Reconciliation logic

Given the full remote set (from `query`) and the full declared set
(`items`):

1. **Match** — pair remote and declared items using `key`
2. **Missing** — declared item has no matching remote item → fire `missing`
   with the declared item as body
3. **Drift** — matched pair, fields in declared item differ from remote →
   fire `found` with the declared item as body
4. **Converged** — matched pair, all declared fields match → noop
5. **Orphan** — remote item has no matching declared item → fire `orphan`
   with `orphan_state` as body

Only keys present in the declared item are compared during drift detection.
Extra fields in the remote response are ignored — same semantics as
[`rest.resource`]({{< relref "resource#state-and-drift-detection" >}}).

## Bindings

Bindings work exactly like [`rest.resource` bindings]({{< relref
"resource#bindings" >}}), but resolve **per-item** against the matched
remote object. Each `found` or `orphan` request gets its own set of
resolved bindings.

```scampi
found    = rest.request { method = "PUT", path = "/api/users/{id}" }
bindings = {"id": rest.jq { expr = "._id" }}
```

## Orphan handling

When `orphan` is set, remote items not present in the declared set trigger
the orphan request. The `orphan_state` dict is sent as the JSON body —
useful for soft-removal patterns like clearing a flag:

```scampi
orphan       = rest.request { method = "PUT", path = "/api/users/{id}" }
orphan_state = {"use_fixedip": false, "fixed_ip": ""}
```

For hard deletion, use a DELETE request with no body:

```scampi
orphan = rest.request { method = "DELETE", path = "/api/users/{id}" }
```

When `orphan` is omitted, extra remote items are left alone. This gives
additive-only behavior — you can add and update items without touching
anything else.

## Examples

### Additive only (no orphan cleanup)

```scampi
rest.resource_set {
  desc  = "DNS records"
  query = rest.request {
    method = "GET"
    path   = "/api/v1/zones/example.com/records"
    check  = rest.jq { expr = ".[]" }
  }
  key     = rest.jq { expr = ".name" }
  items   = [
    {"name": "app",  "type": "A", "value": "198.51.100.5"},
    {"name": "mail", "type": "A", "value": "198.51.100.10"},
  ]
  missing = rest.request { method = "POST", path = "/api/v1/zones/example.com/records" }
  found   = rest.request { method = "PUT",  path = "/api/v1/zones/example.com/records/{id}" }
  bindings = {"id": rest.jq { expr = ".id" }}
}
```

### Full set reconciliation (with orphan cleanup)

```scampi
rest.resource_set {
  desc  = "fixed IP clients"
  query = rest.request {
    method = "GET"
    path   = "/api/s/default/rest/user"
    check  = rest.jq { expr = ".data[] | select(.use_fixedip == true)" }
  }
  key   = rest.jq { expr = ".mac" }
  items = [
    {"mac": "aa:bb:cc:dd:ee:01", "name": "server1", "fixed_ip": "10.0.0.10", "use_fixedip": true},
    {"mac": "aa:bb:cc:dd:ee:02", "name": "server2", "fixed_ip": "10.0.0.11", "use_fixedip": true},
  ]
  missing      = rest.request { method = "POST", path = "/api/s/default/rest/user" }
  found        = rest.request { method = "PUT",  path = "/api/s/default/rest/user/{id}" }
  orphan       = rest.request { method = "PUT",  path = "/api/s/default/rest/user/{id}" }
  bindings     = {"id": rest.jq { expr = "._id" }}
  orphan_state = {"use_fixedip": false, "fixed_ip": ""}
}
```
