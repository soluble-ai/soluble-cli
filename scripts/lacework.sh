#!/bin/bash

# This script simulates what happens when the lacework CLI invokes
# the soluble CLI.

set -euo pipefail

export LW_ACCOUNT=$(set -e; lacework configure show account)
export LW_API_KEY=$(set -e;lacework configure show api_key)
export LW_API_SECRET=$(set -e; lacework configure show api_secret)
export LW_API_TOKEN=$(set -e; lacework access-token)
export LW_COMPONENT_NAME=iac

exec go run main.go "$@"