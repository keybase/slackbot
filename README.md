## Keybase Chat build bot

[![Build Status](https://github.com/keybase/slackbot/actions/workflows/ci.yml/badge.svg)](https://github.com/keybase/slackbot/actions)
[![GoDoc](https://godoc.org/github.com/keybase/slackbot?status.svg)](https://godoc.org/github.com/keybase/slackbot)

The project provides build bots controlled from a configured Keybase Chat
conversation. Set `KEYBASE_CHAT_CONVID` and either run an existing Keybase
client (`KEYBASE_LOCATION`/`KEYBASE_HOME`) or configure
`KEYBASE_ONESHOT_USERNAME` and `KEYBASE_ONESHOT_PAPERKEY`.

Then install and run a bot, and post `!examplebot help` in the configured
conversation.
