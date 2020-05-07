# Soluble CLI

This is the command line interface for [Soluble](https://soluble.ai).

## Installation

On MacOS use [homebrew](https://brew.sh):

    brew install soluble-ai/soluble/soluble-cli

To upgrade to the latest version:

    brew upgrade soluble-ai/soluble/soluble-cli

At the moment, other platforms must be installed from source.  Assuming you have [go](https://golang.org/) installed:

    git checkout https://github.com/soluble-ai/soluble-cli.git
    ./hack/build.sh
    ./soluble version

## Usage

You'll need to generate an access token from the [UI](https://app.soluble.cloud/admin/tokens/access).  Copy the access token and run:

    soluble auth set-access-token --acces-token <your-access-token>

If successful the CLI will show your user profile.

Some useful commands:

    # get help
    soluble help
    # deploy an agent
    soluble agent deploy
    # list clusters
    soluble cluster list
    # list queries
    soluble query list
    # run a query
    soluble query run --query-name deployments

