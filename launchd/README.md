## Launchd

Launchd plists and scripts for running from the bot.

### Installing a launchd plist

- Copy plist to ~/Library/LaunchAgents
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
