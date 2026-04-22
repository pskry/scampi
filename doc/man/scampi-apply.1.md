% SCAMPI-APPLY 1 "" scampi

# NAME

scampi-apply - converge the system to the desired state

# SYNOPSIS

**scampi apply** \[**--only** *blocks*\] \[**--targets** *targets*\] *config*

# DESCRIPTION

Reads a declarative configuration file and executes the required operations
to converge the system to the desired state.

The command is idempotent: running it multiple times only applies changes
when the current state differs from the declared state. Each operation
follows a Check/Execute pattern — it inspects the target first, then only
mutates what needs to change.

Actions execute sequentially. Within each action, operations run in parallel
according to their dependency graph.

# OPTIONS

**--only** *blocks*
  Filter to specific deploy blocks (comma-separated). Only steps within
  the named blocks are executed.

**--targets** *targets*
  Filter to specific targets (comma-separated). Only steps targeting the
  named hosts are executed.

# EXAMPLES

Apply the full configuration:

    $ scampi apply site.scampi

Apply only the production deploy block:

    $ scampi apply --only prod site.scampi

Apply to a single target while developing:

    $ scampi apply --targets web1 site.scampi

Dry-run first, then apply:

    $ scampi check site.scampi
    $ scampi apply site.scampi

# SEE ALSO

**scampi**(1), **scampi-plan**(1), **scampi-check**(1),
**scampi-inspect**(1)
