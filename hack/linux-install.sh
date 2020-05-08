#!/bin/sh

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
