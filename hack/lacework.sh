#!/bin/bash

# This script simulates what happens when the lacework CLI invokes
# the soluble CLI.

set -euo pipefail

export LW_ACCOUNT=$(lacework configure show account)
export LW_API_KEY=$(lacework configure show api_key)
export LW_API_SECRET=$(lacework configure show api_secret)
export LW_API_TOKEN=$(lacework access-token)
export LW_COMPONENT_NAME=iac

exec go run main.go "$@"