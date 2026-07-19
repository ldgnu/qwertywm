#!/bin/sh
quote=$(curl -s --max-time 2 "https://api.quotable.io/random" 2>/dev/null | jq -r '.content + " — " + .author' 2>/dev/null)
[ -z "$quote" ] && quote="$(fastfetch --pipe 2>/dev/null | head -1 | sed 's/^[[:space:]]*//')"
[ -z "$quote" ] && quote="CachyOS"
echo "{\"text\":\"${quote}\"}"
