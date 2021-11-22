#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

client_dir="$dir/../../client"
echo "Loading release tool"
(cd "$client_dir/go/buildtools"; go install "github.com/keybase/release")
release_bin="$GOPATH/bin/release"

url=`"$release_bin" save-log --bucket-name=$BUCKET_NAME --path="$READ_PATH"`
"$dir/send.sh" "Log saved to $url"
