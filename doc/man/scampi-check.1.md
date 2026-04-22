% SCAMPI-CHECK 1 "" scampi

# NAME

scampi-check - check the current system state against a configuration

# SYNOPSIS

**scampi check** \[**--only** *blocks*\] \[**--targets** *targets*\] *config*

# DESCRIPTION

Reads a declarative configuration file and inspects the target system to
determine which operations are already satisfied and which would need to
execute.

No changes are made to the system. Unlike **plan**, this command evaluates
the actual system state — it connects to targets and runs checks, but
never mutates anything.

The output uses semantic colors: green for operations already satisfied,
yellow for operations that would change something, and red for failures.

# OPTIONS

**--only** *blocks*
  Filter to specific deploy blocks (comma-separated).

**--targets** *targets*
  Filter to specific targets (comma-separated).

# EXAMPLES

Check the full configuration:

    $ scampi check site.scampi

Check with verbose output to see why each op passed or failed:

    $ scampi check -v site.scampi

Check only the database target:

    $ scampi check --targets db site.scampi

# SEE ALSO

**scampi**(1), **scampi-plan**(1), **scampi-apply**(1)
