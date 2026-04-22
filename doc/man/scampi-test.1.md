% SCAMPI-TEST 1 "" scampi

# NAME

scampi-test - run scampi test files

# SYNOPSIS

**scampi test** \[*path*\]

# DESCRIPTION

Runs scampi test files. Test files are scampi configurations with the
*_test.scampi* suffix that use the **test.\*** builtins to declare
expected state and verify behavior.

The *path* argument controls which tests to run:

- No argument: runs *\*_test.scampi* in the current directory
- A directory: runs *\*_test.scampi* in that directory
- A path ending in **/...**: runs tests recursively
- A file path: runs that specific test file

The command exits with status 1 if any test fails.

# EXAMPLES

Run tests in the current directory:

    $ scampi test

Run tests recursively from the project root:

    $ scampi test ./...

Run tests in a specific directory:

    $ scampi test tests/

Run a single test file:

    $ scampi test nginx_test.scampi

# SEE ALSO

**scampi**(1)
