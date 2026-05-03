---
title: scampi vs ansible — deploy bench
linkTitle: Bench
description: Reproducible scampi-vs-ansible time-to-converge numbers on a representative deploy workflow. Methodology and reproducer included; numbers are a snapshot.
draft: true
---

A reproducible benchmark for the speed claim. Methodology, both
configurations, the harness, and the raw numbers all live in
[`bench/`](https://codeberg.org/scampi-dev/scampi/src/branch/main/bench)
in the repo. Anyone can spin up 3 hosts and rerun.

## Scenario

Deploy nginx with a representative real-world configuration to **3
fresh hosts**:

- One main `nginx.conf` rendered from a template
- Three vhost configs in `/etc/nginx/conf.d/` (default, site A, site B)
- One landing page in `/var/www/html/`
- Service started and enabled, reloaded if any config template changed

That's 5 templates + a package install + a service unit per host,
matched feature-for-feature between scampi and ansible. The two
configurations live side-by-side under `bench/scampi/deploy.scampi`
and `bench/ansible/site.yml` — neither is handicapped, both are
idiomatic for their tool.

## Pinning

Everything that affects the result is pinned and recorded:

- **Debian template** `debian-12-standard_12.12-1_amd64.tar.zst`
- **nginx package** `nginx=1.22.1-9+deb12u4`
- **ansible-core** `2.18.5` in a venv from
  `bench/ansible/requirements.txt`
- **scampi** version recorded per run via `scampi version` (see
  `<timestamp>.metadata.txt`)
- **hyperfine** version recorded the same way
- **Container resources** 2 vCPU, 2 GB RAM, 8 GB disk per LXC

If anything moves between runs, the recorded metadata exposes it.

## Methodology

- **3× LXC containers** on a Proxmox VE homelab host, ZFS-backed
  storage, virtio-net on a Linux bridge. Specs above.
- **Controller**: TBD (laptop / a 4th VM — disclosed when numbers land).
- **Network**: LAN, sub-millisecond RTT to the targets. Differences
  will be more pronounced over WAN where per-op connect cost
  dominates.
- **Ansible config**: `forks = 10`, `pipelining = true`,
  `ControlMaster=auto` with 60s `ControlPersist` — matches what
  scampi does at the protocol level (one TCP per host, multiplexed
  channels). Without ControlMaster, ansible re-dials per task and
  the comparison reflects ansible-without-tuning rather than the
  tool itself.
- **Harness**: `hyperfine`. N runs each (default 10), reports
  mean / median / stddev / min / max / outliers.
- **Cold vs warm**:
  - **Cold**: hyperfine `--prepare` rolls every LXC back to a
    `pristine` ZFS snapshot before each timed run (uncounted). Each
    cold timing measures a true from-scratch deploy. Snapshot rollback
    is byte-identical state, unlike `apt purge && reinstall` which
    leaves apt's cached metadata warm.
  - **Warm**: one untimed `--warmup 1` run, then N timed re-runs
    against converged state. Each warm timing measures the
    idempotent / no-op path.

## Results

> **Pending.** Numbers will land here once the bench has been run
> against the published methodology, twice: first against the
> pre-release `./build/bin/scampi`, then against the alpha.8 release
> binary installed via `curl get.scampi.dev | sh`. The reproducer is
> committed; if the numbers in the post and the numbers you measure
> don't match, the methodology page lists the controller and network
> specifics for the published run.

When numbers are in, this section will report **median** and **p95**
for each (tool × phase) combination, plus links to the raw JSON in
`bench/results/published/`.

If the baseline wins on any metric, that gets reported too. The point
is to be right, not to win.

## Reproduce

```bash
# 1. Provision 3 LXCs + bootstrap perf-user + pristine snapshot (one shot)
scampi apply bench/provision.scampi

# 2. Set up the ansible venv with pinned ansible-core
python3 -m venv bench/ansible/venv
./bench/ansible/venv/bin/pip install -r bench/ansible/requirements.txt

# 3. Copy + edit the inventory
cp bench/ansible/inventory.ini.example bench/ansible/inventory.ini

# 4. Build scampi (or install via `curl get.scampi.dev | sh` post alpha.8)
just build

# 5. Run both tools, capture per-run timings
bash bench/run.sh

# 6. Tear down
scampi apply bench/teardown.scampi
```

`bench/README.md` has the full prerequisite list, env-var tunables,
and the pinning policy.

## Caveats

- LAN-local results don't generalise directly to WAN deployments.
- Idiomatic ansible includes `ControlMaster` because that's how anyone
  actually using ansible at scale runs it. Numbers without
  `ControlMaster` would penalise ansible for a tuning gap, not for the
  tool's design — that wouldn't be a fair comparison.
- Container guests are smaller than typical production VMs;
  apt/file-system work might be marginally faster than on a busy host.
- These are **snapshot** numbers, not a continuous benchmark. If you
  re-run on different infra or after either tool has shipped a new
  release, expect different absolute values; the *delta* should
  remain in the same direction.
