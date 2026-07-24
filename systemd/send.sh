#! /usr/bin/env bash

set -e -u -o pipefail

repo_root="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
exec "$repo_root/scripts/send.sh" "$@"
