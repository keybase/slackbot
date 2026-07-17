#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

client_dir="$dir/../../client"
logpath=${LOG_PATH:-}
label=${LABEL:-}
nolog=${NOLOG:-""} # Don't show log at end of job
bucket_name=${BUCKET_NAME:-"prerelease.keybase.io"}
: ${SCRIPT_PATH:?"Need to set SCRIPT_PATH to run script"}

echo "Loading release tool"
(cd "$client_dir/go/buildtools"; go install "github.com/keybase/client/go/release")
release_bin="$GOPATH/bin/release"

err_report() {
  url=`$release_bin save-log --bucket-name=$bucket_name --path=$logpath --noerr`
  "$dir/send.sh" "Error \`$label\`, see $url"
}

trap 'err_report $LINENO' ERR

"$SCRIPT_PATH"

# Record the commit built by a successful automated build so the bot can skip
# rebuilding it on the next timed run.
record_commit=${AUTOMATED_BUILD_COMMIT:-""}
record_commit_path=${AUTOMATED_BUILD_COMMIT_PATH:-""}
if [ -n "$record_commit" ] && [ -n "$record_commit_path" ]; then
  echo "$record_commit" > "$record_commit_path"
fi

if [ "$nolog" = "" ]; then
  url=`$release_bin save-log --bucket-name=$bucket_name --path=$logpath --noerr`
  "$dir/send.sh" "Finished \`$label\`, view log at $url"
fi
