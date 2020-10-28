#!/bin/bash

cd "$(dirname ${0})/src"
for dir in ./*; do
  if ! [[ -d "${dir}" ]]; then
    continue
  fi
  tar czvf "../${dir}.tar.gz" "${dir}"
done
