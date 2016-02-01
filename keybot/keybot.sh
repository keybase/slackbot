#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

git pull --ff-only
go get -u github.com/keybase/slackbot/keybot
../send/send.sh "Keybot starting"
go run $GOPATH/src/github.com/keybase/slackbot/keybot/main.go
