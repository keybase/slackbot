# `--ignore-pause` flag for keybot

## Problem

Pausing the bot (`!keybot pause`) blocks every command except `resume` and
`config`. Operators sometimes need to run a single one-off command while the
bot is paused — today that requires resuming, running the command, and
re-pausing, which opens a window where queued/automated commands can run.

## Goal

`!keybot build darwin --ignore-pause` (or any keybot command with the flag)
runs exactly once while the bot stays paused. The paused state is never
mutated.

## Scope

keybot only. winbot and tuxbot are unchanged: the shared central gate will let
a `--ignore-pause` invocation through, but their kingpin parsers reject the
unknown flag and return a usage error, so nothing executes and the bot stays
paused.

## Design

Pause is enforced at two layers; both learn about the flag:

1. **Central gate** (`bot.go` `RunCommand`, line 121). Today:

   ```go
   if args[0] != "resume" && args[0] != "config" && b.Config().Paused() {
   ```

   Add a helper that scans args for the literal `--ignore-pause` token; when
   present, the gate does not block. The flag is left in args so the extension
   parser sees it.

2. **keybot extension** (`keybot/keybot.go` + `keybot/main.go`).
   - App-level kingpin flag: `ignorePause := app.Flag("ignore-pause", "Run this command even when the bot is paused").Bool()` — app-level so it applies to every subcommand.
   - `runScript` gains an `ignorePause bool` parameter. Its pause early-return
     becomes `if bot.Config().Paused() && !ignorePause { ... }`. All call
     sites pass `*ignorePause`.

`SetPaused` is never called; there is no temporary-unpause race. Dry-run
behavior is unchanged and still takes precedence over the pause check in
`runScript`.

Side effect (accepted): with the flag, non-script keybot subcommands like
`cancel` also work while paused, which is desirable for one-off operations.

## Testing

- `bot_test.go`: paused bot blocks a command without the flag; with
  `--ignore-pause` the command runs and `Paused()` remains true.
- `keybot/main_test.go`: keybot `Run` with `--ignore-pause` on a paused
  (non-dry-run is not testable here; use dry-run path) parses cleanly and does
  not error.
