#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

client_dir="$GOPATH/src/github.com/keybase/client"

echo "Loading release tool"
"$client_dir/packaging/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

if [ -n "$SMOKETEST_BUILD_A" ] && [ -n "$SMOKETEST_PLATFORM" ] && [ -n "$SMOKETEST_BUILD_ENABLE" ]; then
  "$release_bin" set-build-in-testing --build-a="$SMOKETEST_BUILD_A" --platform="$SMOKETEST_PLATFORM" --enable="$SMOKETEST_BUILD_ENABLE"
  "$client_dir/packaging/slack/send.sh" "Successfully set enable to $SMOKETEST_BUILD_ENABLE for release $SMOKETEST_BUILD_A."
else
  "$client_dir/packaging/slack/send.sh" "Error: Missing environment variables in smoketest command."
fi
