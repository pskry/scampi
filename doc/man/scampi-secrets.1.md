% SCAMPI-SECRETS 1 "" scampi

# NAME

scampi-secrets - manage age-encrypted secrets

# SYNOPSIS

**scampi secrets** \[**-i** *identity*\] *subcommand* \[*args*\]

# DESCRIPTION

Manages age-encrypted secrets for use in scampi configurations. Secrets
are stored as a JSON file where each value is independently encrypted
using age (https://age-encryption.org/).

Every developer generates their own age keypair with **secrets init**.
The private key stays on their machine; the public key is shared with
the team. When setting a secret, it's encrypted to the setter's key
plus any additional recipients specified with **-r**.

## Secrets file resolution

1. **--file** / **-f** flag
2. **SCAMPI_SECRETS_FILE** environment variable
3. *secrets.age.json* in the current directory
4. *secrets.file.json* in the current directory

## Identity resolution

1. **-i** / **--identity** flag on the **secrets** parent command
2. **SCAMPI_AGE_KEY** environment variable (raw private key)
3. **SCAMPI_AGE_KEY_FILE** environment variable (path to key file)
4. *~/.config/scampi/age.key*

# PARENT OPTIONS

**-i**, **--identity** *path*
  Path to age identity file. Overrides default resolution.

# SUBCOMMANDS

## init

Generate an age keypair. The private key is saved to
*~/.config/scampi/age.key* and the public key is printed to stdout.

**--force**
  Overwrite an existing key file (requires interactive confirmation).

## pubkey

Print the public key for the current age identity.

## set

Encrypt and store a secret value.

**scampi secrets set** *key* \[*value*\]

If *value* is omitted, reads from stdin. Use **-r** to add recipients
beyond your own key. Use **-f** to specify the secrets file.

**-f**, **--file** *path*
  Path to secrets file.

**-r**, **--recipient** *pubkey*
  Additional age recipient public key. Repeatable.

## get

Decrypt and print a secret value. Without a key argument, lists all
available keys.

**scampi secrets get** \[*key*\]

**-f**, **--file** *path*
  Path to secrets file.

## list

List available secret keys.

**-f**, **--file** *path*
  Path to secrets file.

## del

Remove a secret from the store.

**scampi secrets del** *key*

**-f**, **--file** *path*
  Path to secrets file.

## info

Show secret keys and their recipient counts.

**-f**, **--file** *path*
  Path to secrets file.

## recrypt

Re-encrypt all secrets with the current identity and specified
recipients. This is how you add or revoke recipients.

**scampi secrets recrypt** \[**-r** *pubkey*\]...

**-f**, **--file** *path*
  Path to secrets file.

**-r**, **--recipient** *pubkey*
  Age recipient public key. Repeatable. Replaces the existing recipient
  list — only the keys specified here (plus your own) will be able to
  decrypt. Running without **-r** drops all other recipients (requires
  interactive confirmation).

# EXAMPLES

Generate a keypair:

    $ scampi secrets init
    age1abc123...

Store a secret with an additional recipient (e.g. CI pipeline):

    $ scampi secrets set db.password s3cret -r age1ci...

Read a secret:

    $ scampi secrets get db.password

Pipe a secret from a file:

    $ cat token.txt | scampi secrets set api.token

List all secrets and their recipient counts:

    $ scampi secrets info

Add a new team member as a recipient to all secrets:

    $ scampi secrets recrypt -r age1alice... -r age1ci...

Revoke a recipient (re-encrypt for remaining keys only):

    $ scampi secrets recrypt -r age1alice...

Use a different identity for a project:

    $ scampi secrets -i keys/prod.key get db.password

# ENVIRONMENT

**SCAMPI_SECRETS_FILE**
  Path to the secrets file.

**SCAMPI_AGE_KEY**
  Raw age private key string.

**SCAMPI_AGE_KEY_FILE**
  Path to an age key file.

# FILES

*secrets.age.json*
  Default secrets store in the working directory. Created on first
  **scampi secrets set**.

*~/.config/scampi/age.key*
  Default age identity file.

# SEE ALSO

**scampi**(1), **age**(1)
