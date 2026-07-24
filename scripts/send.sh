#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

json_escape() {
  local input=$1
  local output=""
  local char
  local escaped
  local code
  local i
  local LC_ALL=C

  for ((i = 0; i < ${#input}; i++)); do
    char=${input:i:1}
    case "$char" in
      '"')
        output="${output}\\\""
        ;;
      \\)
        output="${output}\\\\"
        ;;
      $'\b')
        output="${output}\\b"
        ;;
      $'\f')
        output="${output}\\f"
        ;;
      $'\n')
        output="${output}\\n"
        ;;
      $'\r')
        output="${output}\\r"
        ;;
      $'\t')
        output="${output}\\t"
        ;;
      *)
        printf -v code '%d' "'$char"
        if ((code < 0)); then
          code=$((code + 256))
        fi
        if ((code < 32)); then
          printf -v escaped '\\u%04x' "$code"
          output="${output}${escaped}"
        else
          output="${output}${char}"
        fi
        ;;
    esac
  done

  printf '%s' "$output"
}

# send to keybase chat if we have it in the environment
convid=${KEYBASE_CHAT_CONVID:-}
if [ -n "$convid" ]; then
  echo "Sending to Keybase convID: $convid"
  location=${KEYBASE_LOCATION:-"keybase"}
  home=${KEYBASE_HOME:-$HOME}
  body="$*"
  escaped_convid=$(json_escape "$convid")
  escaped_body=$(json_escape "$body")
  payload="{\"method\":\"send\",\"params\":{\"options\":{\"conversation_id\":\"${escaped_convid}\",\"message\":{\"body\":\"${escaped_body}\"}}}}"
  "$location" --home "$home" chat api -m "$payload"
fi
