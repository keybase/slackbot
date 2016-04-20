### Instructions for getting the Keybase buildbot running on Linux

TODO: I should script this. Unfortunately there are "systemctl --user"
issues that make this difficult. Can I get around this by manually
setting `DBUS_SESSION_BUS_ADDRESS` after the `enable-linger` call?
(Further note, `XDG_RUNTIME_DIR` also works for this.)

- Create an account called "keybasebuild": `sudo useradd -m
  keybasebuild`
- Add user "keybasebuild" to the "docker" group: `sudo gpasswd -a
  keybasebuild docker`.
- Get an SSH key with access to the keybase GitHub repos and the keybase
  account on aur.archlinux.org.
- Clone several keybase repos into /home/keybasebuild:
  - client
  - kbfs
  - kbfs-beta
  - server-ops
  - slackbot
  - `git clone aur@aur.archlinux.org:keybase-git $(mktemp -d)` (This
    repo is a throwaway, but it makes sure you have the right SSH keys,
    and it adds the host key to ~/.ssh/known_hosts.)
- Import the code signing PGP secret key. After import, remove the
  password from this key. (TODO: Something more interesting with
  yubikeys.)
- Set up s3cmd (~/.s3cfg) with credentials for
  s3://prerelease.keybase.io.
- Create /home/keybasebuild/keybot.env with the following lines:

    ```
    SLACK_TOKEN=<slack token here>
    SLACK_CHANNEL=bot
    ```

- Start/enable-on-boot the systemd service files.
  - `sudo loginctl enable-linger keybasebuild`
  - With a proper login session as keybasebuild (SSH is one way):
    - `mkdir -p ~/.config/systemd/user`
    - `cp ~/slackbot/systemd/keybase.*.{service,timer} ~/.config/systemd/user/`
    - `systemctl --user enable --now keybase.keybot.service`
    - `systemctl --user enable --now keybase.buildplease.timer`
- Take the bot out of dry-run mode by messaging `!tuxbot toggle-dryrun`.
  (This assumes that the `SLACK_TOKEN` you defined above corresponds to
  the `tuxbot` Slack user.)
