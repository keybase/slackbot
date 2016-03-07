## Launchd

Launchd plists and scripts for running from the bot.

### Installing a launchd plist

- Copy plist to ~/Library/LaunchAgents
- Set the necessary environment variables (for tokens, etc)
- Run `launchctl load -w file.plist`
