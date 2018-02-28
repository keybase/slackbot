#! /usr/bin/env bash

# This script cleans the go path, go-ios path, go-android path
set -e -u -o pipefail

cd "$GOPATH/src/client"
echo $(git clean -f)
echo $(git checkout master)

cd "$GOPATH/../go-ios/src/client"
echo $(git clean -f)
echo $(git checkout master)

cd "$GOPATH/../go-android/src/client"
echo $(git clean -f)
echo $(git checkout master)
