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

tag="${1:-}"

if [ -z "$tag" ]; then
  echo "usage: $0 release-tag"
  exit 1
fi

formula=/usr/local/Homebrew/Library/Taps/soluble-ai/homebrew-soluble/soluble-cli.rb
formula_dir=$(dirname $formula)

if ! git -C $formula_dir diff --quiet; then
  echo "$formula_dir is dirty, stash or restore the tap before proceeding"
  exit 1
fi

(
  cd $formula_dir
  git checkout master
  git pull
)

releases=https://github.com/soluble-ai/soluble-cli/releases
tarball="https://github.com/soluble-ai/soluble-cli/archive/${tag}.tar.gz"
echo "Getting $tarball"
curl --fail -L -o target/source.tar.gz -H "Accept:application/octet-stream" "$tarball"
hash=$(shasum -a 256 target/source.tar.gz | awk '{print $1}')

sed -I "" -e "s/sha256 .*/sha256 \"$hash\"/" -e "s,url .*,url \"$tarball\"," $formula
(
  cd $formula_dir
  git diff
  echo "# Test with:"
  echo "  brew upgrade soluble-ai/soluble/soluble-cli"
  echo "# Commit with:"
  echo "  cd $(pwd)"
  echo "  git commit -a -m 'Version $tag'"
  echo "  git push"
)