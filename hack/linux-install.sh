#!/bin/sh

if wget 2> /dev/null; then
http_get () {
    wget --header "$1" -q "$2"
}
http_download () {
    wget --header "$1" "$2"
}
else
http_get () {
    echo $2
    curl -L -H "$1" -s "$2"
}
http_download () {
    curl -L "$2"
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
http_download "Accept:application/octet-stream" \
    "$tarball" | tar zxvf - soluble
chmod a+rx soluble
./soluble version
