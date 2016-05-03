#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

logpath=${LOG_PATH:-}

client_dir="$GOPATH/src/github.com/keybase/client"
bucket_name="prerelease.keybase.io"
platform="darwin"

echo "Loading release tool"
"$client_dir/packaging/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

err_report() {
  "$client_dir/packaging/slack/send.sh" "Error see $logpath"
}

trap 'err_report $LINENO' ERR


if [ -n "$RELEASE_TO_PROMOTE" ];
then
  "$release_bin" promote-a-release --release="$RELEASE_TO_PROMOTE" --bucket-name="$bucket_name" --platform="$platform"
  "$client_dir/packaging/slack/send.sh" "Promoted $platform release $RELEASE_TO_PROMOTE ($bucket_name)"
else
  "$release_bin" promote-releases --bucket-name="$bucket_name" --platform="$platform"
  "$client_dir/packaging/slack/send.sh" "Promoted $platform release on ($bucket_name)"
fi
