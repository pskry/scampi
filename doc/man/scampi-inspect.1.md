% SCAMPI-INSPECT 1 "" scampi

# NAME

scampi-inspect - show resolved state or diff file content

# SYNOPSIS

**scampi inspect** \[**--only** *blocks*\] \[**--targets** *targets*\] *config*

**scampi inspect** **--diff** \[**-i**\] *config* \[*path*\]

# DESCRIPTION

Reads a declarative configuration file and shows the resolved state of all
steps after evaluation.

In its default mode, **inspect** lists every step with its resolved field
values — useful for verifying that variable interpolation, conditionals,
and module composition produced the expected result.

With **--diff**, the command compares file content managed by scampi against
the current state on the target:

- **--diff** alone lists all diffable file paths
- **--diff** *path* diffs that specific file
- **--diff -i** opens a fuzzy finder to pick a file interactively

The diff tool is selected from **SCAMPI_DIFFTOOL**, **DIFFTOOL**, **EDITOR**,
or falls back to **diff**(1).

# OPTIONS

**--only** *blocks*
  Filter to specific deploy blocks (comma-separated).

**--targets** *targets*
  Filter to specific targets (comma-separated).

**--diff**
  Switch to diff mode. Without a path argument, lists all diffable files.
  With a path, diffs that file against the target's current content.

**-i**, **--interactive**
  Pick a file interactively using the program set in
  **SCAMPI_FUZZY_FINDER** (e.g. **fzf**, **sk**). Requires **--diff**.

# EXAMPLES

Show resolved state for all steps:

    $ scampi inspect site.scampi

List all files that can be diffed:

    $ scampi inspect site.scampi --diff

Diff a specific managed file:

    $ scampi inspect site.scampi --diff /etc/nginx/nginx.conf

Pick a file to diff with fzf:

    $ export SCAMPI_FUZZY_FINDER=fzf
    $ scampi inspect site.scampi --diff -i

Pipe diffable paths to a script:

    $ scampi inspect site.scampi --diff | xargs -I{} scampi inspect site.scampi --diff {}

# ENVIRONMENT

**SCAMPI_DIFFTOOL**
  Diff tool to use. Falls back to **DIFFTOOL**, then **EDITOR**, then
  **diff**(1).

**SCAMPI_FUZZY_FINDER**
  Fuzzy finder for interactive mode (e.g. **fzf**, **sk**).

# SEE ALSO

**scampi**(1), **scampi-check**(1), **scampi-apply**(1)
