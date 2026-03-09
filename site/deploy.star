target.ssh(name="web", host="scampi.dev", user="deploy")

deploy(
    name = "site",
    targets = ["web"],
    steps = [
        pkg(packages=["caddy"], state="present"),
        dir(path="/var/www/scampi.dev", perm="0755"),
        template(
            src = "./templates/Caddyfile.tmpl",
            dest = "/etc/caddy/Caddyfile",
            perm = "0644",
            owner = "root",
            group = "root",
        ),
        copy(
            src = "../build/site-dist/site.tar.gz",
            dest = "/tmp/scampi-site.tar.gz",
            perm = "0644",
            owner = "root",
            group = "root",
        ),
        run(
            desc = "extract site content",
            check = "test /var/www/scampi.dev/index.html -nt /tmp/scampi-site.tar.gz",
            apply = "tar xzf /tmp/scampi-site.tar.gz -C /var/www/scampi.dev",
        ),
        service(name="caddy", state="running", enabled=True),
    ],
)
