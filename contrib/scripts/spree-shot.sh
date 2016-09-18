#!/usr/bin/env bash
set -e
set -x

spree_endpoint=spree.roemmich.org:4285

screenshot_file=$(echo "/tmp/$(date +"%F")_$(date +"%N").png")
gnome-screenshot -f "$screenshot_file" -a
convert "${screenshot_file}" -quality 75 "${screenshot_file}"
display_path=$(~/bin/spreectl --ca.cert.file=/home/r2/bin/spree.ca.crt --key.file=/home/r2/bin/spreectl.key --cert.file=/home/r2/bin/spreectl.crt -rpc.addr=${spree_endpoint} upload -src="${screenshot_file}" -file="${screenshot_file}" | jq -r '.shot.path' | tr -d '\n')
display_url=https://spree.roemmich.org${display_path}
notify-send -t 3000 -a spree "screenshot saved to clipboard: ${display_url}"
echo -n ${display_url} | xclip -i
