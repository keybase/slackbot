#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

script="${SCRIPT_TO_RUN:-}"
if [ -z "$script" ] ; then
  echo "run_and_send_stdout needs a script argument."
  exit 1
fi

result=$($script)

"$dir/send.sh" "\`$script\`:\`\`\`$result\`\`\`"
