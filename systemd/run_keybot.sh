#! /usr/bin/env bash

set -e -u -o pipefail

cd "$(dirname "$BASH_SOURCE")/../tuxbot"

export GOPATH="$(pwd)/gopath"

if ! [ -e "$GOPATH" ] ; then
  # Build the local GOPATH.
  mkdir -p "$GOPATH/src/github.com/keybase"
  ln -s "$(git rev-parse --show-toplevel)" gopath/src/github.com/keybase/slackbot
fi

go install github.com/keybase/slackbot
go install github.com/keybase/slackbot/tuxbot

# Wait for the network.
while ! ping -c 3 slack.com ; do
  sleep 1
done

exec "$GOPATH/bin/tuxbot"
