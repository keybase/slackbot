#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

gopath=${GOPATH:-}
logpath=${LOG_PATH:-}
: ${SCRIPT_PATH:?"Need to set SCRIPT_PATH to run script"}


if [ "$gopath" = "" ]; then
  echo "No GOPATH"
  exit 1
fi

client_dir="$gopath/src/github.com/keybase/client"

echo "Loading release tool"
"$client_dir/packaging/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"


err_report() {
  url=`$release_bin save-log --bucket-name=prerelease.keybase.io --path=$logpath --noerr`
  "$client_dir/packaging/slack/send.sh" "Error see $url"
}

trap 'err_report $LINENO' ERR

"$SCRIPT_PATH"
