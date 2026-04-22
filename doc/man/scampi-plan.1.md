% SCAMPI-PLAN 1 "" scampi

# NAME

scampi-plan - show the execution plan for a configuration

# SYNOPSIS

**scampi plan** \[**--only** *blocks*\] \[**--targets** *targets*\] *config*

# DESCRIPTION

Reads a declarative configuration file and prints the execution plan
without applying any changes.

The plan shows the operations that would be executed by **scampi apply**,
but does not inspect or modify the target system. Use this to verify
the structure of your configuration before running **check** or **apply**.

# OPTIONS

**--only** *blocks*
  Filter to specific deploy blocks (comma-separated).

**--targets** *targets*
  Filter to specific targets (comma-separated).

# EXAMPLES

Preview the full plan:

    $ scampi plan site.scampi

Plan only the staging block:

    $ scampi plan --only staging site.scampi

Plan with verbose output to see operation details:

    $ scampi plan -vv site.scampi

# SEE ALSO

**scampi**(1), **scampi-check**(1), **scampi-apply**(1)
