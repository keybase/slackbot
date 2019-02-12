If you want to run a single build by hand, see the README in
https://github.com/keybase/client: packaging/linux/README.md

### Instructions for getting the Keybase buildbot running on Linux

- Create an account called "keybasebuild": `sudo useradd -m keybasebuild`
  - NOTE: If you use a different name, you will need to tweak the
    *.service files in this directory. They hardcode paths that include
    the username.
- Add user "keybasebuild" to the "docker" group: `sudo gpasswd -a keybasebuild docker`
- Allow the keybasebuild account to start systemd services on boot: `sudo loginctl enable-linger keybasebuild`
- Do a *real log in* as that user. That means either a graphical
  desktop, or via SSH. In particular, if you try to `sudo` into this
  user, several steps below will fail.
- Install all the credentials you need. We have a sepate "build-linux"
  repo for this with its own README-- ask Max where it is.
- Clone three repos into /home/keybasebuild:
  - https://github.com/keybase/client
  - https://github.com/keybase/kbfs
  - https://github.com/keybase/slackbot (this repo)
- Enable the systemd service files. (These are the commands that will
  fail if you don't have a real login.)
  - `mkdir -p ~/.config/systemd/user`
  - `cp ~/slackbot/systemd/keybase.*.{service,timer} ~/.config/systemd/user/`
  - `systemctl --user enable --now keybase.keybot.service`
  - `systemctl --user enable --now keybase.buildplease.timer`
- Take the bot out of dry-run mode by messaging `!tuxbot toggle-dryrun`
  in the #bot channel.

For stathat logging, add a `STATHAT_EZKEY` env variable to the envfile used by the unit.
