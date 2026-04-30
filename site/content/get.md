---
title: ""
linkTitle: Install
---

<style>
/* Kill spacing from blank title */
#content > br, #content > h1 { display: none; }
#content > .content { margin-top: 0; }
#content > div.hx\:mb-16 { margin-bottom: 1.9rem; }
/* Mascot in left margin */
.get-mascot {
  position: sticky;
  top: 6rem;
  float: left;
  margin-left: -200px;
  width: 160px;
}
@media (max-width: 1280px) { .get-mascot { display: none; } }
</style>

<img src="/scampi-get.png" alt="scampi mascot" class="get-mascot">

{{< hextra/hero-badge link="https://codeberg.org/scampi-dev/scampi/releases" >}}
  <span>All releases</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

{{< hextra/hero-headline >}}
  Get scampi
{{< /hextra/hero-headline >}}

{{< hextra/hero-subtitle >}}
  Pick your method. Be converging in seconds.
{{< /hextra/hero-subtitle >}}

## One-liner

```bash
curl get.scampi.dev | sh
```

Downloads the latest release of **both `scampi` and `scampls`** (the LSP
server), verifies the release SSH signature on `SHA256SUMS`, checks each
binary's SHA256, and installs to `~/.local/bin` (or `/usr/local/bin` if
it doesn't exist).

**Just the CLI** (e.g. CI runners):

```bash
curl get.scampi.dev/cli | sh
```

**Just the LSP:**

```bash
curl get.scampi.dev/lsp | sh
```

**Custom path:**

```bash
curl get.scampi.dev | sh -s -- -o ~/.local/bin
```

Supported platforms: Linux, macOS, and FreeBSD (amd64/arm64).

## Go

```bash
go install scampi.dev/scampi/cmd/scampi@latest
go install scampi.dev/scampi/cmd/scampls@latest
```

Requires Go {{< go-version >}}+.

## Manual download

Prebuilt binaries for all supported platforms are available on
[Codeberg releases](https://codeberg.org/scampi-dev/scampi/releases).

Download the binary for your platform, verify against `SHA256SUMS` (and ideally
the [signature](#verify-a-release) too), and place it on your `PATH`.

## Verify a release

Releases are signed with an Ed25519 SSH key (`releases@scampi.dev`). Each
release ships `SHA256SUMS` plus `SHA256SUMS.sig` — the install one-liner
verifies both automatically. To verify by hand:

```bash
# 1. Pick a release tag
TAG=v0.1.0-alpha.7

# 2. Download SHA256SUMS, signature, and the binary you want
curl -fLO "https://codeberg.org/scampi-dev/scampi/releases/download/${TAG}/SHA256SUMS"
curl -fLO "https://codeberg.org/scampi-dev/scampi/releases/download/${TAG}/SHA256SUMS.sig"
curl -fLO "https://codeberg.org/scampi-dev/scampi/releases/download/${TAG}/scampi-linux-amd64"

# 3. Verify the signature on SHA256SUMS
cat > allowed_signers <<'EOF'
releases@scampi.dev ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEDBbJOSWyfk9kJhHjUmSJVIax9lxGnOjwpL4dSheQfu
EOF

ssh-keygen -Y verify \
  -f allowed_signers \
  -I releases@scampi.dev \
  -n file \
  -s SHA256SUMS.sig \
  < SHA256SUMS

# 4. Verify the binary against SHA256SUMS
sha256sum --ignore-missing -c SHA256SUMS
```

If the signature verification fails, **don't run the binary**. File a
report via the [security policy](https://codeberg.org/scampi-dev/scampi/src/branch/main/SECURITY.md) —
it could mean the release was tampered with, or that the key was rotated
and your `allowed_signers` line is stale.

The signing pubkey is also embedded in the
[install.sh](https://codeberg.org/scampi-dev/scampi/src/branch/main/scripts/install.sh)
source if you want to grab it from there.

## Build from source

```bash
git clone https://codeberg.org/scampi-dev/scampi.git
cd scampi
just build
```

Produces `./build/bin/scampi` and `./build/bin/scampls`. Requires Go {{< go-version >}}+ and [just](https://just.systems).
