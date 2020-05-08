#!/bin/bash

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
    content_type="application/gzip"
    if [ "${f##.}" = "zip" ]; then
        content_type="application/zip"
    fi
    gc --fail -X POST -d @$f -H "Content-Type: $content_type" "$url" | jq .
done
