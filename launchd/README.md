## Launchd

Launchd plists and scripts for running from the bot.

### Installing a launchd plist

- Copy plist to ~/Library/LaunchAgents
- Set the necessary environment variables (for tokens, etc)
- Edit the plist to set the necessary environment variables (for tokens, etc)
- Load the plist
```
launchctl load -w file.plist
```

After editing, you need to reload the plist:

```
launchctl unload file.plist
launchctl load -w file.plist
```

### Dependencies

If you are installing a plist that has dependencies be sure to set them up for
the GOPATH used in the plist.

```
GOPATH=/Users/test/go go get -u github.com/keybase/slackbot
GOPATH=/Users/test/go-ios go get -u github.com/keybase/slackbot
GOPATH=/Users/test/go-android go get -u github.com/keybase/slackbot

GOPATH=/Users/test/go-ios go get golang.org/x/mobile/cmd/gomobile
GOPATH=/Users/test/go-ios /Users/test/go-ios/bin/gomobile init

GOPATH=/Users/test/go-android go get golang.org/x/mobile/cmd/gomobile
GOPATH=/Users/test/go-android /Users/test/go-android/bin/gomobile init
```
