% SCAMPI-GEN 1 "" scampi

# NAME

scampi-gen - generate scampi modules from external schemas

# SYNOPSIS

**scampi gen api** \[**-o** *output*\] \[**-p** *prefix*\] \[**-n** *name-prefix*\] \[**-m** *module*\] \[**--no-test**\] *spec.yaml*

# DESCRIPTION

Generates scampi modules from external schema formats. Currently supports
OpenAPI specifications.

## gen api

Reads an OpenAPI specification (YAML or JSON) and generates a *.api.scampi*
module with typed functions for each API endpoint. By default, a companion
smoke test file (*_test.scampi*) is generated alongside.

The output file name is derived from the spec filename unless **-o** is given.
Use **-o -** to write to stdout.

# OPTIONS

**-o**, **--output** *path*
  Output file path. Defaults to the spec filename with an *.api.scampi*
  extension. Use **-** for stdout.

**-p**, **--prefix** *path*
  Path prefix prepended to all generated routes (e.g. */integration*).

**-n**, **--name-prefix** *prefix*
  Prefix prepended to all generated function names (e.g. *legacy_*).

**-m**, **--module** *name*
  Override the module declaration name. Defaults to a name derived from
  the spec filename.

**--no-test**
  Skip generating the companion smoke test file.

# EXAMPLES

Generate from an OpenAPI spec:

    $ scampi gen api openapi.yaml

Generate with a route prefix and custom output path:

    $ scampi gen api -p /v2 -o api/billing.api.scampi billing-spec.yaml

Generate with a name prefix for namespacing:

    $ scampi gen api -n legacy_ old-api.yaml

Write to stdout for inspection:

    $ scampi gen api -o - openapi.yaml

Skip test generation:

    $ scampi gen api --no-test openapi.yaml

# SEE ALSO

**scampi**(1), **scampi-mod**(1)
