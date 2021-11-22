#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

# send to keybase chat if we have it in the environment
convid=${KEYBASE_CHAT_CONVID:-}
if [ -n "$convid" ]; then
  echo "Sending to Keybase convID: $convid"
  location=${KEYBASE_LOCATION:-"keybase"}
  home=${KEYBASE_HOME:-$HOME}
  $location --home $home chat api -m "{\"method\":\"send\", \"params\": {\"options\": { \"conversation_id\": \"$convid\" , \"message\": { \"body\": \"$@\" }}}}"
fi

go install "github.com/keybase/slackbot"
go install "github.com/keybase/slackbot/send"
send_bin="$GOPATH/bin/send"

"$send_bin" -i=1 "$@"
