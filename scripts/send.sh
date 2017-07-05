#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

"$dir/goinstall.sh" "github.com/keybase/slackbot"

go install "github.com/keybase/slackbot/send"
send_bin="$GOPATH/bin/send"

"$send_bin" -i=1 "$@"
