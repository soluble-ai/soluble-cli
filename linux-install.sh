#!/bin/sh
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


if curl --help > /dev/null 2>&1 ; then
http_get () {
    echo $2
    curl -L -H "$1" -s "$2"
}
http_download () {
    curl -L "$2"
}
else
http_get () {
    wget --header "$1" -q "$2"
}
http_download () {
    wget -O - --header "$1" "$2"
}
fi

releases=https://github.com/soluble-ai/soluble-cli/releases
tag="$1"
if [  -z "$tag" ]; then
    tag=$(http_get "Accept:application/json" "$releases/latest" | \
        grep tag_name | sed 's/.*tag_name":"//' | sed 's/",.*//')
fi
if [ -z "$tag" ]; then
    echo "Cannot find a release to install"
    exit 1
fi
tarball="$releases/download/${tag}/soluble_${tag}_linux_amd64.tar.gz"
echo "Getting $tarball"
rm -f soluble
http_download "Accept:application/octet-stream" \
    "$tarball" | tar zxf - soluble
if [ -f soluble ]; then
    echo "Downloaded soluble executable"
    chmod a+rx soluble
    ls -l ./soluble
    ./soluble version
else
    echo "Download failed"
    exit 1
fi
