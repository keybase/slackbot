#! /usr/bin/env bash

# This script answers
#   1) Does repo X exist?
#   2) What's the diff?

set -e -u -o pipefail

addGOPATHPrefix="${PREFIX_GOPATH:-}"
repo="${REPO:-}"

if [ -z "$repo" ] ; then
  echo "git_diff.sh needs a repo argument."
  exit 1
fi

if [ -n "$addGOPATHPrefix" ] ; then
  repo="$GOPATH/src/$repo"
fi

if [ ! -d "$repo" ] ; then
  echo "Repo directory '$repo' does not exist."
  exit 1
fi

cd "$repo"

if [ ! -d ".git" ] ; then
  # This intentionally doesn't support bare repos. Some callers are going to
  # want to mess with the working copy.
  echo "Directory '$repo' is not a git repo."
  exit 1
fi

current_status="$(git status --porcelain)"
if [ -n "$current_status" ] ; then
  echo "Repo '$repo' isn't clean."
  echo "$current_status"
  git diff
else
  echo "Repo '$repo' is clean."
fi
