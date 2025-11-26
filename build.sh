#!/bin/bash

GOBIN="go"

$GOBIN version

LDFLAGS="-s -w -checklinkname=0"
ROOT=${PWD}
OUTPUT="${ROOT}/dist/TorrServer"

#### Build web
echo "Build web"
export NODE_OPTIONS=--openssl-legacy-provider
$GOBIN run gen_web.go

#### Build server
echo "Build server"
cd "${ROOT}/server" || exit 1
$GOBIN clean -i -r -cache # --modcache
$GOBIN mod tidy

BUILD_FLAGS=(-ldflags "${LDFLAGS}" -tags nosqlite -trimpath)
GOOS=linux GOARCH=amd64 ${GOBIN} build "${BUILD_FLAGS[@]}" -o "${OUTPUT}-linux-amd64" ./cmd

