#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

git pull --ff-only
go install github.com/keybase/slackbot/keybot
"$GOPATH/bin/keybot"
