#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

name=${NAME:-}

# NAME comes from the slack command, so it's a good idea to completely
# whitelist the package names here.
if [ "$name" = "go" ]; then
  brew upgrade go
elif [ "$name" = "yarn" ]; then
  brew upgrade yarn
elif [ "$name" = "fastlane" ]; then
  gem update fastlane
  gem cleanup
fi
