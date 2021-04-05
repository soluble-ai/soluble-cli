#!/bin/bash
# Copyright 2021 Soluble Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -euo pipefail

green() {
    echo -e "\033[32m${@}\033[0m" >&2
}

run() {
    green soluble "${@}"
    go run main.go "${@}"
}

export SOLUBLE_OPTS=--force-color

run version
run auth profile --format none
# we don't have any secrets here other than in testdata, so the --error-not-empty will
# fail right away
run secrets-scan --exclude go.sum --exclude 'pkg/**/testdata/*.json' \
  --exclude 'pkg/tools/cloudsploit/**' --error-not-empty --upload
run auto-scan --upload --image nginx:1.19 --skip secrets --exclude 'pkg/inventory/testdata/k/t/*.yaml'

if [ -n "${SOLUBLE_API_TOKEN:-}" -a -n "${GITHUB_ACTIONS:-}" ]; then
#    temporarily commenting it out
#    run build update-pr
fi
