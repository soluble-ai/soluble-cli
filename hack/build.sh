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

VERSION=$(git describe --tags --dirty --always)

echo "Version ${VERSION}"

build_time=$(date -u +%Y-%m-%dT%H:%M:%S+00:00)

ldflags="-ldflags=-X 'github.com/soluble-ai/soluble-cli/pkg/version.Version=${VERSION}' \
-X 'github.com/soluble-ai/soluble-cli/pkg/version.BuildTime=${build_time}'"

set -e

echo "Running go test"
go test -cover ./...

linter=golangci-lint
if [ -x ./bin/golangci-lint ]; then
    linter=./bin/golangci-lint
fi

if "${linter}" --help > /dev/null 2>&1; then
    echo "Running ${linter}"
    "${linter}" run -E stylecheck -E gosec -E goimports -E misspell -E gocritic \
      -E whitespace -E goprintffuncname \
      -e G402 ; # we turn off TLS verification by option
else
    echo "golangci-lint not available, skipping lint"
fi

echo "Running go generate"
go generate ./...

rm -rf dist
mkdir -p dist

IFS=" "

for p in "linux amd64 tar" "windows amd64 zip .exe" "darwin amd64 tar"; do
    read -a os_arch <<< "$p"
    echo "Building $VERSION for ${os_arch[0]} ${os_arch[1]}"
    rm -rf target
    mkdir target
    # need to specify osusergo,netgo tags to actually get a static
    # binary - thanks https://www.arp242.net/static-go.html
    #
    # -trimpath was added to go 1.13 (our minimum build target)
    # which ultimately supports reproducible binary build by
    # removing otherwise hardcoded filesystem paths in the binary.
    set -x
    GOOS=${os_arch[0]} GOARCH=${os_arch[1]} \
        go build -o target/soluble${os_arch[3]} -tags ci,osusergo,netgo -trimpath "$ldflags"
    { set +x; } 2> /dev/null
    cp LICENSE README.md target
    pkg=${os_arch[2]}
    name=soluble_${VERSION}_${os_arch[0]}_${os_arch[1]}
    echo "Packaging $name"
    (
        cd target
        if [ "$pkg" = "tar" ]; then
            tar vcf - * | gzip -9 > ../dist/$name.tar.gz
        elif [ "$pkg" = "zip" ]; then
            zip ../dist/$name.zip *
        fi
    )
done

ls -l dist/*
