#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

client_dir="$dir/../../client"
echo "Loading release tool"
(cd "$client_dir/go/buildtools"; go install "github.com/keybase/client/go/release")
release_bin="$GOPATH/bin/release"

"$release_bin" broken-release --release="$BROKEN_RELEASE" --bucket-name="$BUCKET_NAME" --platform="$PLATFORM"
"$dir/send.sh" "Removed $BROKEN_RELEASE for $PLATFORM ($BUCKET_NAME)"
