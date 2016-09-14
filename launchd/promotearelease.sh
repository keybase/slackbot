#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

bucket_name=${BUCKET_NAME:-"prerelease.keybase.io"}
client_dir="$GOPATH/src/github.com/keybase/client"

echo "Loading release tool"
"$client_dir/packaging/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

if [ -n "$RELEASE_TO_PROMOTE" ]; then
  "$release_bin" promote-a-release --release="$RELEASE_TO_PROMOTE" --bucket-name="$BUCKET_NAME" --platform="$PLATFORM"
  "$client_dir/packaging/slack/send.sh" "Promoted $PLATFORM release $RELEASE_TO_PROMOTE ($BUCKET_NAME)"
else
  "$release_bin" promote-releases --bucket-name="$BUCKET_NAME" --platform="$PLATFORM"
  "$client_dir/packaging/slack/send.sh" "Promoted $PLATFORM release on ($BUCKET_NAME)"
fi
