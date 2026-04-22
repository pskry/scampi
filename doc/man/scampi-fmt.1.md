% SCAMPI-FMT 1 "" scampi

# NAME

scampi-fmt - format scampi configuration files

# SYNOPSIS

**scampi fmt** \[**-l**\] *files-or-directories...*

# DESCRIPTION

Formats scampi configuration files to the canonical style. Accepts file
paths, directory paths (formats all *.scampi* files in that directory), or
paths ending in **/...** for recursive formatting.

Test files (*_test.scampi*) are excluded from formatting.

Without **-l**, reformatted files are written in place and their paths
are printed to stdout. With **-l**, files that *would* change are listed
but not modified, and the command exits with status 1 if any would change.

# OPTIONS

**-l**, **--list**
  List files that would be reformatted without writing them. Exits with
  status 1 if any files need formatting (useful in CI).

# EXAMPLES

Format all scampi files in the current directory:

    $ scampi fmt .

Format recursively from the project root:

    $ scampi fmt ./...

Check formatting in CI (non-zero exit if anything would change):

    $ scampi fmt -l ./...

Format a single file:

    $ scampi fmt config.scampi

# SEE ALSO

**scampi**(1)
