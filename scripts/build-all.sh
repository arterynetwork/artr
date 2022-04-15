#!/usr/bin/env bash

apk add --no-cache git make
git config --global --add safe.directory /art-node
cd /art-node
rm -rf builds

for os in linux windows darwin
do
  CGO_ENABLED=0 GOARCH=amd64 GOOS=$os OUTPUT_DIR=builds/${GOOS}_${GOARCH} GOFLAGS="-trimpath -o=${OUTPUT_DIR}/" /bin/sh -c 'mkdir -p ${OUTPUT_DIR} && make build'
done
