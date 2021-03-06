#!/bin/bash
# Copyright 2020 Soluble Inc
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


# Attach dist/* binaries to release

gc() {
    green Running curl "$@"
    curl -H "Authorization: Bearer $GITHUB_TOKEN" "$@"
}

green() {
    echo -e "\033[32m${@}\033[0m" >&2
}

set -e
set -o pipefail

if [ "$GITHUB_EVENT_NAME" != "release" -o "$GITHUB_TOKEN" = "" ]; then
    green "Skipping dist"
    exit 0
fi

if ! find dist -type f -name '*.tar.gz' > /dev/null 2>&1; then
    green "No dist files present, skipping dist"
    exit 0
fi

# GITHUB_REF should be the tag of the release in the form refs/tags/vx.y.z
# see https://help.github.com/en/actions/reference/events-that-trigger-workflows#release-event-release

tag=$(echo $GITHUB_REF | sed 's,.*/v,v,')

if [[ ! $tag =~ ^v[0-9] ]]; then
    green "Tag is not a release tag, skipping dist"
    exit 0
fi

upload_url=$(gc -s https://api.github.com/repos/soluble-ai/soluble-cli/releases/tags/$tag | \
    jq  -r .upload_url | sed 's/{.*//')

green "Upload url for release is $upload_url"

# upload a release asset
# https://developer.github.com/v3/repos/releases/#upload-a-release-asset

for f in dist/*; do
    green "Uploading $f"
    url="${upload_url}?name=$(basename $f)"
    gc --fail -X POST --data-binary @$f -H "Content-Type: application/octet-stream" "$url" | jq .
done
