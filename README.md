> **Note**
> Source code releases of the soluble-cli are no longer available.
> Binary release distributions will continue to be published here. Refer to the installation guide below.

# Lacework IAC CLI

This is the command line interface for [Lacework IAC Security](https://docs.lacework.com/iac/).

## Installation

On MacOS use [homebrew](https://brew.sh):

    brew update
    brew install --cask soluble-ai/soluble/soluble

To upgrade to the latest version:

    brew upgrade --cask soluble-ai/soluble/soluble

On linux, run:

    wget -O - https://raw.githubusercontent.com/soluble-ai/soluble-cli/master/linux-install.sh | sh

Or:

    curl https://raw.githubusercontent.com/soluble-ai/soluble-cli/master/linux-install.sh | sh

The install will drop the executable in the current directory.  If you run this as `root` or can sudo to root,
the install will try to move the binary to `/usr/local/bin/soluble`.

Windows executables can be found on the releases page.

## Run Security Scans

First, you'll need to signup with Lacework IAC:

    soluble login

Now you can run security scans on your code:

    # scan terraform IAC
    soluble terraform-scan -d ~/my-stuff
    # scan for secrets
    soluble secrets-scan -d ~/my-stuff
    # scan kubernetes manifests
    soluble kubernetes-scan -d ~/my-stuff
    # scan Helm charts
    soluble helm-scan -d ~/my-stuff

See https://docs.lacework.com/iac/ for more information.

## Run As Component

Follow the installation steps from Lacework CLI to configure it
https://docs.lacework.net/cli/

Run the below commands to get a release version of soluble cli running as IaC component under the lacework cli.
```
lacework component install iac

# verify the installation of iac component
lacework component list

# verify the iac configuration as Component
lacework iac config show
```

## Run As Component with local changes

Follow the installation steps from Lacework CLI to configure it
https://docs.lacework.net/cli/

Run the script to test local changes
```
./scripts/lacework.sh tf-scan 
```
