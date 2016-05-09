#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

client_dir="$GOPATH/src/github.com/keybase/client"
bucket_name="prerelease.keybase.io"

echo "Loading release tool"
"$client_dir/packaging/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

"$release_bin" broken-release --release="$BROKEN_RELEASE" --bucket-name="$bucket_name"
"$client_dir/packaging/slack/send.sh" "Broken $BROKEN_RELEASE ($bucket_name)"

report=`"$release_bin" updates-report --bucket-name="$bucket_name"`
"$client_dir/packaging/slack/send.sh" "\`\`\`$report\`\`\`"
