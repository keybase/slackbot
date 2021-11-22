#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

client_dir="$dir/../../client"
echo "Loading release tool"
(cd "$client_dir/go/buildtools"; go install "github.com/keybase/release")
release_bin="$GOPATH/bin/release"

dryrun=""
if [ $DRY_RUN == 'true' ]; then
  dryrun="--dry-run"
fi

if [ -n "$RELEASE_TO_PROMOTE" ]; then
  "$release_bin" promote-a-release --release="$RELEASE_TO_PROMOTE" --bucket-name="$BUCKET_NAME" --platform="$PLATFORM" $dryrun
  "$dir/send.sh" "Promoted $PLATFORM release $RELEASE_TO_PROMOTE ($BUCKET_NAME)"
else
  if [ $DRY_RUN == 'true' ]; then
    "$dir/send.sh" "Can't dry-run without a specific release to promote"
    exit 1
  fi
  "$release_bin" promote-releases --bucket-name="$BUCKET_NAME" --platform="$PLATFORM"
  "$dir/send.sh" "Promoted $PLATFORM release on ($BUCKET_NAME)"
fi
