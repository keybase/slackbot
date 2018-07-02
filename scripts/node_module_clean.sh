#! /usr/bin/env bash

# This script cleans the go-ios node_modules 
set -e -u -o pipefail

rm -rf "$GOPATH/../go-ios/src/github.com/keybase/client/shared/node_modules"
rm -rf "$GOPATH/../go-android/src/github.com/keybase/client/shared/node_modules"
cd "$GOPATH/../go-android/src/github.com/keybase/client/shared"
yarn rn-packager-wipe-cache
