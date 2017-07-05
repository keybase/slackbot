#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

echo "Loading release tool"
"$dir/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

"$release_bin" set-build-in-testing --build-a="$SMOKETEST_BUILD_A" --platform="$PLATFORM" --enable="$SMOKETEST_ENABLE" --max-testers="$SMOKETEST_MAX_TESTERS"
"$dir/../send/send.sh" "Successfully set enable to $SMOKETEST_ENABLE for release $SMOKETEST_BUILD_A."
