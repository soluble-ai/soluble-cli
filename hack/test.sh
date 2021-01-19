#!/bin/bash

set -euo pipefail

green() {
    echo -e "\033[32m${@}\033[0m" >&2
}

run() {
    green soluble $*
    go run main.go ${@}
}

export SOLUBLE_OPTS=--force-color

run version
run auth profile --format none

if [ -n "${SOLUBLE_API_TOKEN:-}" -a -n "${GITHUB_ACTIONS:-}" ]; then
    run auto-scan --upload --image nginx:1.19 --skip secrets --exclude 'pkg/inventory/testdata/k/t/*.yaml'
    # we don't have any secrets here, so the --error-not-empty will
    # fail right away
    run secrets-scan --exclude go.sum --exclude 'pkg/tools/secrets/testdata/results.json' --error-not-empty --upload
    run build update-pr
fi
