# Soluble CLI

This is the command line interface for [Soluble](https://soluble.ai).

## Installation

On MacOS use [homebrew](https://brew.sh):

    brew install soluble-ai/soluble/soluble-cli

To upgrade to the latest version:

    brew upgrade soluble-ai/soluble/soluble-cli

On linux, run:

    wget -O - https://raw.githubusercontent.com/soluble-ai/soluble-cli/master/linux-install.sh | sh

Or:

    curl https://raw.githubusercontent.com/soluble-ai/soluble-cli/master/linux-install.sh | sh

The install will drop the executable in the current directory.  If you run this as `root` or can sudo to root,
the install will try to move the binary to `/usr/local/bin/soluble`.

Windows executables can be found on the releases page.

## Run Security Scans

Run security scans on your code with:

    # scan terraform IAC
    soluble terraform-scan -d ~/my-stuff
    # scan for secrets
    soluble secrets-scan -d ~/my-stuff
    # scan container images
    soluble image-scan -d ~/my-stuff
    # scan kubernetes manifests
    soluble kubernetes-scan -d ~/my-stuff
    # scan Helm charts
    soluble helm-scan -d ~/my-stuff

If you'd like to manage the findings of those tools with [Soluble](https://app.soluble.cloud), you'll have to authenticate the CLI with:

    soluble login

Then re-run the scan with with `--upload` flag, as in:

    soluble terraform-scan -d ~/my-stuff --upload

Some of the scans support multiple tools.  For example, `soluble terraform-scan` by default scans [terraform files](https://www.terraform.io/) with [checkov](https://github.com/bridgecrewio/checkov), and `soluble terraform-scan tfsec` scans with [tfsec](https://github.com/tfsec/tfsec).

Use the builtin help e.g. `soluble help terraform-scan` to see the supported scanners and options.
