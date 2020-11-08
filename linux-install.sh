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

soluble_exe=soluble

if [ "$(uname -s)" != "Linux" ]; then
    echo "ERROR: not linux"
    exit 1
fi

if curl --help > /dev/null 2>&1 ; then
http_get () {
    echo $2
    curl -L -H "$1" -s "$2"
}
http_download () {
    curl -sL "$2"
}
else
http_get () {
    wget -O - --header "$1" -q "$2"
}
http_download () {
    wget -O - --header "$1" "$2"
}
fi

configure_cli () {

    if [ -f "${HOME}/.soluble/cli-config.json" ]; then
        echo "soluble config (${HOME}/.soluble/cli-config.json already exists"
    else 
        if [ "${SOLUBLE_API_URL}" != "" ]; then
            echo "Setting SOLUBLE_API_URL=${SOLUBLE_API_URL}"
            ${soluble_exe} config  set APIServer "${SOLUBLE_API_URL}" >/dev/null
        fi
        if [ "${SOLUBLE_API_TOKEN}" != "" ]; then
            echo "Setting SOLUBLE_TOKEN=*******"
            ${soluble_exe} config  set APIToken "${SOLUBLE_API_TOKEN}" >/dev/null
        fi
        if [ "${SOLUBLE_ORG_ID}" != "" ]; then
            echo "Setting SOLUBLE_ORG_ID=${SOLUBLE_ORG_ID}"
            ${soluble_exe} config  set Organization "${SOLUBLE_ORG_ID}" >/dev/null
        fi
    fi
}

install_cli () {
    if [ "$(whoami)" = "root" ]; then
        mv ./soluble /usr/local/bin/soluble
    else 
        if [ "$(sudo -n whoami)" = "root" ]; then
            sudo -n mv ./soluble /usr/local/bin/soluble
        else
            echo "Could not install to /usr/local/bin/soluble"
        fi
    fi

    if [ -x "/usr/local/bin/soluble" ]; then
        soluble_exe=/usr/local/bin/soluble
    fi
}

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

    install_cli
    configure_cli

else
    echo "Download failed"
    exit 1
fi
