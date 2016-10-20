#!/usr/bin/env bash

# this is pretty much tied to my setup, which has gdate, jq, and imagemagick installed via homebrew
set -e
set -x

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
spree_endpoint=spree.roemmich.org:4285
screenshot_file=$(echo "/tmp/$(/uar/local/bin/gdate +"%F")_$(/usr/local/bin/gdate +"%N").png")
screencapture -o -i ${screenshot_file}
/usr/local/bin/convert "${screenshot_file}" -quality 75 "${screenshot_file}"
display_path=$(~/bin/spreectl --ca.cert.file=${DIR}/spree.ca.crt --key.file=${DIR}/spreectl.key --cert.file=${DIR}/spreectl.crt -rpc.addr=${spree_endpoint} upload -src="${screenshot_file}" -file="${screenshot_file}" | /usr/local/bin/jq -r '.shot.path' | tr -d '\n')
display_url="https://spree.roemmich.org${display_path}"
osascript -e "display notification \"screenshot saved to clipboard: ${display_url}\" with title \"Spree\""
echo -n ${display_url} | pbcopy
