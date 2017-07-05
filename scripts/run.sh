#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

logpath=${LOG_PATH:-}
label=${LABEL:-}
nolog=${NOLOG:-""} # Don't show log at end of job
bucket_name=${BUCKET_NAME:-"prerelease.keybase.io"}
: ${SCRIPT_PATH:?"Need to set SCRIPT_PATH to run script"}

echo "Loading release tool"
"$dir/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

err_report() {
  url=`$release_bin save-log --bucket-name=$bucket_name --path=$logpath --noerr`
  "$dir/send.sh" "Error \`$label\`, see $url"
}

trap 'err_report $LINENO' ERR

"$SCRIPT_PATH"

if [ "$nolog" = "" ]; then
  url=`$release_bin save-log --bucket-name=$bucket_name --path=$logpath --noerr`
  "$dir/send.sh" "Finished \`$label\`, view log at $url"
fi
