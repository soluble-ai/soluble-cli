# Soluble CLI

This is the command line interface for [Soluble](https://soluble.ai).

## Installation

On MacOS use [homebrew](https://brew.sh):

    brew install soluble-ai/soluble/soluble-cli

To upgrade to the latest version:

    brew upgrade soluble-ai/soluble/soluble-cli

On linux, run:

    wget -O - https://raw.githubusercontent.com/soluble-ai/soluble-cli/master/linux-install.sh | sh
    # or
    curl https://raw.githubusercontent.com/soluble-ai/soluble-cli/master/linux-install.sh | sh

The install will drop the executable in the current directory.  If you run this as `root` or can sudo to root,
the install will try to move the binary to `/usr/local/bin/soluble`.

Windows executables can be found on the releases page.

## Scan Infrastructure-As-Code

Find infrastructure-as-code files with:

    soluble iac-inventory local -d ~/my-stuff

This will search under `~/my-stuff` for a variety of infrastructre-as-code files and print out the results.

Those files can be scanned with running:

    soluble iac-scan all -d ~/my-stuff --print-tool-results

If you'd like to manage the findings of those tool with [Soluble](https://app.soluble.cloud), you'll have to authenticate the CLI with:

    soluble login

Then re-run the scan with:

    soluble iac-scan all -d ~/my-stuff --upload

## CI Integration

(WIP - basic notes)

1. Set the environment variable `SOLUBLE_API_TOKEN` to an API token from https://app.soluble.cloud/admin/tokens/access 
2. Add a step to run `soluble iac-scan all -d . --upload`, or run individual scans (see `soluble iac-scan --help`)
3. At then end of your CI job, run `soluble build-report`

## Build from source

Assuming you have [go](https://golang.org/) installed:

    git checkout https://github.com/soluble-ai/soluble-cli.git
    ./hack/build.sh
    ./soluble version
