% SCAMPLS 1 "" scampi

# NAME

scampls - Language Server Protocol server for scampi

# SYNOPSIS

**scampls** \[**--log** *file*\]

**scampls check** *file...*

**scampls hover** *file* *line* *col*

**scampls def** *file* *line* *col*

**scampls scan** *file*

# DESCRIPTION

**scampls** is an LSP server for scampi configuration files. It
communicates over stdin/stdout using the standard JSON-RPC transport and
provides diagnostics, completion, hover, goto-definition, and signature
help.

When invoked without a subcommand, it starts the LSP server in stdio
mode — this is the mode your editor uses.

The subcommands are CLI tools for debugging the LSP pipeline without
needing an editor.

# OPTIONS

**--log** *file*
  Write debug log to *file*. Useful for diagnosing LSP issues.

# SUBCOMMANDS

## check

Run the diagnostic pipeline on one or more files and print results.
Exits with status 1 if any diagnostics are found.

## hover

Run a hover request at the given position and print the result.
*line* and *col* are 0-indexed.

## def

Run a goto-definition request at the given position and print the
location(s). *line* and *col* are 0-indexed.

## scan

Run hover and definition requests at every cursor position in *file*
to detect crashes. Reports any panics found. Used for robustness
testing.

# EXAMPLES

Start the server (what your editor runs):

    $ scampls

Start with debug logging:

    $ scampls --log /tmp/scampls.log

Check a file for errors from the command line:

    $ scampls check site.scampi

Check multiple files:

    $ scampls check *.scampi

Get hover info at a position:

    $ scampls hover site.scampi 10 5

Find definition at a position:

    $ scampls def site.scampi 10 5

Smoke-test all positions in a file:

    $ scampls scan site.scampi

# EDITOR CONFIGURATION

Most editors auto-detect LSP servers when configured. Point your editor's
LSP client at the **scampls** binary with the filetype set to *scampi*.

For Neovim with nvim-lspconfig:

    vim.lsp.config.scampls = {
      cmd = { 'scampls' },
      filetypes = { 'scampi' },
      root_markers = { 'scampi.mod' },
    }

# SEE ALSO

**scampi**(1)

https://scampi.dev
