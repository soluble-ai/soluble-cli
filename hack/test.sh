#!/bin/bash

set -euo pipefail

green() {
    echo -e "\033[32m${@}\033[0m" >&2
}

run() {
    green soluble $*
    go run main.go ${@}
}

run version
run auth profile --format none

if [ -n "${SOLUBLE_API_TOKEN:-}" -a -n "${GITHUB_ACTIONS:-}" ]; then
    run iac-scan all --upload
    run image-scan --image nginx:1.19 --upload
    run iac-scan build-report
fi
