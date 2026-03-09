---
title: Project Layout
weight: 2
---

Scampi doesn't impose a directory structure, but here are some patterns that work
well as projects grow.

## Minimal

A single file is enough:

```
myproject/
├── deploy.star
└── nginx.conf
```

```python
# deploy.star
target.local(name="dev")

deploy(
    name = "webserver",
    steps = [
        pkg(packages=["nginx"], state="present"),
        copy(src="nginx.conf", dest="/etc/nginx/nginx.conf", perm="0644"),
        service(name="nginx", state="running", enabled=True),
    ],
)
```

```nginx
# nginx.conf
worker_processes auto;

events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        root /var/www/html;
    }
}
```

## Small

Separate targets from deploy logic:

```
myproject/
├── targets.star
├── deploy.star
├── files/
│   ├── nginx.conf
│   └── app.env
└── templates/
    └── Caddyfile.tmpl
```

```python
# targets.star
target.ssh(name="web", host="app.example.com", user="deploy")
```

```python
# deploy.star
load("targets.star", "web")

deploy(
    name = "app",
    targets = ["web"],
    steps = [
        pkg(packages=["caddy"], state="present"),
        copy(src="files/app.env", dest="/opt/app/.env", perm="0600", owner="app", group="app"),
        template(
            src = "templates/Caddyfile.tmpl",
            dest = "/etc/caddy/Caddyfile",
            perm = "0644",
            data = {"domain": "app.example.com"},
        ),
        service(name="caddy", state="running", enabled=True),
    ],
)
```

```
# files/app.env
NODE_ENV=production
PORT=3000
```

```
# templates/Caddyfile.tmpl
{{ .domain }} {
    reverse_proxy localhost:3000
}
```

## Medium

Group by concern when managing multiple services:

```
infra/
├── targets.star
├── web.star
├── db.star
├── monitoring.star
├── files/
│   └── ...
└── templates/
    └── ...
```

```python
# targets.star
target.ssh(name="web", host="web.example.com", user="deploy")
target.ssh(name="db", host="db.example.com", user="deploy")
target.ssh(name="mon", host="mon.example.com", user="deploy")
```

```python
# web.star
load("targets.star", "web")

deploy(
    name = "web",
    targets = ["web"],
    steps = [
        pkg(packages=["nginx", "certbot"], state="present"),
        copy(src="files/nginx.conf", dest="/etc/nginx/nginx.conf", perm="0644"),
        service(name="nginx", state="running", enabled=True),
    ],
)
```

```python
# db.star
load("targets.star", "db")

deploy(
    name = "database",
    targets = ["db"],
    steps = [
        pkg(packages=["postgresql-16"], state="present"),
        copy(src="files/pg_hba.conf", dest="/etc/postgresql/16/main/pg_hba.conf", perm="0640", owner="postgres", group="postgres"),
        service(name="postgresql", state="running", enabled=True),
    ],
)
```

```python
# monitoring.star
load("targets.star", "mon")

deploy(
    name = "monitoring",
    targets = ["mon"],
    steps = [
        pkg(packages=["prometheus", "grafana"], state="present"),
        template(
            src = "templates/prometheus.yml.tmpl",
            dest = "/etc/prometheus/prometheus.yml",
            perm = "0644",
            data = {"targets": ["web.example.com", "db.example.com"]},
        ),
        service(name="prometheus", state="running", enabled=True),
        service(name="grafana-server", state="running", enabled=True),
    ],
)
```

## Large

Split into directories per environment. Use Starlark functions to define steps
once and vary the data per environment:

```
infra/
├── shared/
│   ├── targets.star
│   └── web.star
├── production.star
├── staging.star
└── templates/
    └── nginx.conf.tmpl
```

```python
# shared/targets.star
target.ssh(name="prod-web", host="web.prod.example.com", user="deploy")
target.ssh(name="staging-web", host="web.staging.example.com", user="deploy")
```

```python
# shared/web.star — defines the steps once
def web_steps(domain):
    return [
        pkg(packages=["nginx"], state="present"),
        template(
            src = "templates/nginx.conf.tmpl",
            dest = "/etc/nginx/nginx.conf",
            perm = "0644",
            data = {"values": {"domain": domain, "upstream_port": 3000}},
        ),
        service(name="nginx", state="running", enabled=True),
    ]
```

```python
# production.star — just wiring
load("shared/targets.star", "prod-web")
load("shared/web.star", "web_steps")

deploy(name="prod-web", targets=["prod-web"], steps=web_steps("prod.example.com"))
```

```python
# staging.star — same steps, different values
load("shared/targets.star", "staging-web")
load("shared/web.star", "web_steps")

deploy(name="staging-web", targets=["staging-web"], steps=web_steps("staging.example.com"))
```

The production and staging files are pure wiring — the actual step logic lives in
one place. Use `load()` to share target definitions and step functions across
files.
