#! /usr/bin/env bash

# This script cleans the go path, go-ios path, go-android path
set -e -u -o pipefail

cd "$GOPATH/src/github.com/keybase/client"
echo $(git fetch)
echo $(git clean -f)
echo $(git checkout master)
echo $(git pull)

cd "$GOPATH/../go-ios/src/github.com/keybase/client"
echo $(git fetch)
echo $(git clean -f)
echo $(git checkout master)
echo $(git pull)

cd "$GOPATH/../go-android/src/github.com/keybase/client"
echo $(git fetch)
echo $(git clean -f)
echo $(git checkout master)
echo $(git pull)
