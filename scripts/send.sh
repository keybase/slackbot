#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

"$dir/goinstall.sh" "github.com/keybase/slackbot"

# Outputs to slack if you have slackbot installed and SLACK_TOKEN and
# SLACK_CHANNEL set, otherwise it does nothing (errors are ignored on purpose).

sender="$dir/../send/main.go"

if [ -f $sender ]; then
  go run $sender -i=1 "$@"
else
  echo "[No sender] $@"
fi
