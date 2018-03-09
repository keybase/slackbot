Some random notes:

The android keys are in kbfs so make sure the keybot has keybase running (could move to encrypted git someday)
The bot using launch agents, so look at the plist files in ~/Library/LaunchAgents. When builds kick off it does it through launch agents as well
There are multiple go-paths that exist. The bot runs in ~/go. android builds run from ~/go-android and ios runs from ~/go-ios. The yarn rn-gobuild-* also runs in /tmp like client does
The bot delegates to client's build and publish scripts under packaging so look there too
