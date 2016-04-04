#! /usr/bin/env bash

set -e -u -o pipefail

cd "$(dirname "$BASH_SOURCE")/../keybot"

export GOPATH="$(pwd)/gopath"

if ! [ -e "$GOPATH" ] ; then
  # Build the local GOPATH.
  mkdir -p "$GOPATH/src/github.com/keybase"
  ln -s "$(git rev-parse --show-toplevel)" gopath/src/github.com/keybase/slackbot
fi

go get -v github.com/keybase/slackbot/keybot
go install github.com/keybase/slackbot/keybot

exec "$GOPATH/bin/keybot"
