% SCAMPI 1 "" scampi

# NAME

scampi - declarative system convergence engine

# SYNOPSIS

**scampi** \[**--ascii**\] \[**--color**=*auto*|*always*|*never*\] \[**-v**|**-vv**|**-vvv**\] *command* \[*args*\]

# DESCRIPTION

scampi reads declarative configuration files that describe desired system
state and executes idempotent operations to converge reality to that state.
It works on local machines, remote hosts over SSH, and REST APIs.

Every operation follows a Check/Execute pattern: check whether the current
state already matches the desired state, and only execute when it doesn't.
Running **scampi apply** twice in a row produces no changes the second time.

The typical workflow is **plan** → **check** → **apply**:

- **plan** shows what *would* happen (no target access)
- **check** inspects the live system without modifying it
- **apply** makes the changes

# GLOBAL OPTIONS

**--ascii**
  Force ASCII output. Disables Unicode glyphs.

**--color**=*auto*|*always*|*never*
  Colorize output. Defaults to *auto* (color when writing to a terminal).

**-v**, **-vv**, **-vvv**
  Increase verbosity. **-v** shows *why* decisions were made, **-vv** adds
  *how* operations execute, **-vvv** shows everything.

# COMMANDS

**plan** *config*
  Show the execution plan without touching the target system.

**check** *config*
  Inspect the live system state against the configuration.

**apply** *config*
  Converge the system to the desired state.

**inspect** *config*
  Show resolved state for all steps, or diff file content.

**fmt** *files...*
  Format scampi configuration files.

**test** \[*path*\]
  Run scampi test files.

**mod** *subcommand*
  Manage module dependencies.

**gen** *subcommand*
  Generate scampi modules from external schemas.

**secrets** *subcommand*
  Manage age-encrypted secrets.

**index** \[*step*\]
  List available steps and their documentation.

**legend**
  Show the CLI visual language reference (glyphs and colors).

**version**
  Print the scampi version.

# EXIT CODES

**0**
  Success.

**1**
  User error: invalid configuration, failed plan, or validation error.

**2**
  Internal error or panic (bug).

# ENVIRONMENT

**SCAMPI_SECRETS_FILE**
  Path to the secrets file. Overrides auto-discovery of
  *secrets.age.json* or *secrets.file.json* in the working directory.

**SCAMPI_AGE_KEY**
  Raw age private key string. Used for secret decryption when no
  **--identity** flag is given.

**SCAMPI_AGE_KEY_FILE**
  Path to an age key file. Checked after **SCAMPI_AGE_KEY**.

**SCAMPI_DIFFTOOL**
  Diff tool for **inspect --diff**. Falls back to **DIFFTOOL**, then
  **EDITOR**, then plain **diff**(1).

**SCAMPI_FUZZY_FINDER**
  Fuzzy finder program (e.g. **fzf**, **sk**) for **inspect --diff -i**.

**XDG_CACHE_HOME**
  Module cache location. Defaults to *~/.cache/scampi/mod*.

**XDG_CONFIG_HOME**
  Configuration directory. Defaults to *~/.config*. The age identity
  is stored at *$XDG_CONFIG_HOME/scampi/age.key*.

**SSH_AUTH_SOCK**
  SSH agent socket, forwarded to SSH targets automatically.

# FILES

*scampi.mod*
  Module file in the project root. Lists dependencies and versions.
  Created by **scampi mod init**.

*scampi.sum*
  Checksum file in the project root. Managed automatically alongside
  *scampi.mod*.

*secrets.age.json*
  Default secrets store in the working directory. Created on first
  **scampi secrets set**.

*~/.config/scampi/age.key*
  Default age identity file. Created by **scampi secrets init**.

# EXAMPLES

Plan first, then apply:

    $ scampi plan site.scampi
    $ scampi apply site.scampi

Scope to a single deploy block during development:

    $ scampi apply --only dev site.scampi

Apply to specific targets only:

    $ scampi apply --targets web1,web2 site.scampi

Check what would change, with verbose output:

    $ scampi check -vv site.scampi

Inspect resolved state, then diff a managed file:

    $ scampi inspect site.scampi
    $ scampi inspect site.scampi --diff /etc/nginx/nginx.conf

Interactive diff with fzf:

    $ export SCAMPI_FUZZY_FINDER=fzf
    $ scampi inspect site.scampi --diff -i

# SEE ALSO

**scampi-apply**(1), **scampi-check**(1), **scampi-plan**(1),
**scampi-inspect**(1), **scampi-fmt**(1), **scampi-test**(1),
**scampi-mod**(1), **scampi-gen**(1), **scampi-secrets**(1),
**scampls**(1)

https://scampi.dev
