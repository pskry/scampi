---
title: secrets
weight: 10
---

Manage [age](https://age-encryption.org/)-encrypted secrets. Subcommands
look for the secrets file in this order:

1. `--file` / `-f` flag
2. `$SCAMPI_SECRETS_FILE` environment variable
3. `secrets.age.json` in the current directory
4. `secrets.file.json` in the current directory

The `set` command creates `secrets.age.json` automatically if no file
exists yet — no setup required.

All subcommands that need an age identity resolve it in this order:

1. `-i` / `--identity` flag (on the `secrets` parent command)
2. `$SCAMPI_AGE_KEY` environment variable (raw private key)
3. `$SCAMPI_AGE_KEY_FILE` environment variable (path to key file)
4. `~/.config/scampi/age.key`

Use `--identity` when working with multiple projects that have
different keys (e.g. separate prod and dev identities):

```text
scampi secrets -i keys/prod.key get db.password
scampi secrets -i keys/dev.key set api.url http://localhost:8080
```

## How age encryption works

Every developer runs `scampi secrets init` to generate their own age keypair,
stored in `$XDG_CONFIG_HOME/scampi/` (typically `~/.config/scampi/`). The
private key stays on their machine; the public key is shared with the team.

When you `set` a secret, it's encrypted to your key plus any additional
recipients you specify with `--recipient`. Anyone whose public key was included
as a recipient can decrypt that secret with their private key. Each secret in
the JSON file is encrypted independently, so different secrets can have
different recipient lists.

## Key rotation and revocation

Age bakes the recipient list into the ciphertext. You can't add or remove
recipients without re-encrypting. Use `recrypt` to update the recipient
list for all secrets at once.

**Adding a recipient** (e.g. a CI pipeline):

```text
scampi secrets recrypt -r age1abc...pipeline
```

**Revoking a recipient** (e.g. someone leaving the team): run `recrypt`
with only the recipients you want to keep. The old ciphertext is replaced,
and the revoked key can no longer decrypt the new values.

Always include at least two recipients so a single lost key doesn't make
secrets unrecoverable.

## Disaster recovery

If you lose your private key (e.g. `secrets init --force` overwrites it),
any remaining recipient can re-encrypt the secrets for a new key:

```text
# 1. Generate a new keypair
scampi secrets init

# 2. Get the new public key
scampi secrets pubkey
# age1new...

# 3. Re-encrypt using a surviving recipient's private key
SCAMPI_AGE_KEY="AGE-SECRET-KEY-1SURVIVING..." \
  scampi secrets recrypt \
    -r age1new... \
    -r age1other...
```

This decrypts every secret with the surviving key and re-encrypts for the
new recipient list. The lost key is effectively revoked in the same step.

## init

```text
scampi secrets init [--force]
```

Generate an age keypair for encrypting secrets.

| Flag      | Description                 |
| --------- | --------------------------- |
| `--force` | Overwrite existing key file |

When `--force` is used and a key already exists, an interactive
confirmation prompt is shown. This is blocked entirely in
non-interactive environments (piped input, CI) to prevent accidental
key destruction.

## pubkey

```text
scampi secrets pubkey
```

Print the public key for the current age identity.

## set

```text
scampi secrets set [flags] <key> [value]
```

Encrypt and store a secret value. If `value` is omitted, it is read
from stdin. Empty values are rejected.

Creates the secrets file on first use if it doesn't exist yet.

| Flag                | Description                                      |
| ------------------- | ------------------------------------------------ |
| `-f`, `--file`      | Path to secrets file (default: secrets.age.json) |
| `-r`, `--recipient` | Additional age recipient public key (repeatable) |

## get

```text
scampi secrets get [flags] [key]
```

With a key argument, decrypt and print that secret's value. Without
arguments, list all secret key names in the store.

| Flag           | Description                                      |
| -------------- | ------------------------------------------------ |
| `-f`, `--file` | Path to secrets file (default: secrets.age.json) |

Output is always unformatted, safe for piping:

```text
HOST=$(scampi secrets get vps.host)
```

## del

```text
scampi secrets del [flags] <key>
```

Remove a secret from the store.

| Flag           | Description                                      |
| -------------- | ------------------------------------------------ |
| `-f`, `--file` | Path to secrets file (default: secrets.age.json) |

## info

```text
scampi secrets info [flags]
```

Show all secret keys and how many recipients each value is encrypted
for. Useful for verifying that secrets are encrypted for the expected
number of keys.

| Flag           | Description                                      |
| -------------- | ------------------------------------------------ |
| `-f`, `--file` | Path to secrets file (default: secrets.age.json) |

## recrypt

```text
scampi secrets recrypt [flags]
```

Re-encrypt all secrets with the current identity and the specified
recipients. This replaces the recipient list entirely — any recipient
not included in `-r` flags is removed.

Your own public key (derived from your identity) is always included
automatically.

| Flag                | Description                                      |
| ------------------- | ------------------------------------------------ |
| `-f`, `--file`      | Path to secrets file (default: secrets.age.json) |
| `-r`, `--recipient` | Age recipient public key (repeatable)            |
