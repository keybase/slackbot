#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

client_dir="$GOPATH/src/github.com/keybase/client"

echo "Loading release tool"
"$client_dir/packaging/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

url=`"$release_bin" save-log --bucket-name=prerelease.keybase.io --path="$logpath"`
"$client_dir/packaging/slack/send.sh" "Log saved to $url"
