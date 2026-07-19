#!/bin/sh
data=$(curl -s --max-time 1 http://localhost:10767/playback 2>/dev/null)
if [ -n "$data" ]; then
  track=$(echo "$data" | jq -r '.name // ""' 2>/dev/null)
  artist=$(echo "$data" | jq -r '.artistName // ""' 2>/dev/null)
  state=$(echo "$data" | jq -r '.state // "stopped"' 2>/dev/null)
  if [ "$state" = "playing" ]; then
    icon="▶"
  else
    icon="⏸"
  fi
  if [ -n "$track" ]; then
    echo "{\"text\":\"${icon} ${track} - ${artist}\"}"
    exit
  fi
fi
echo '{"text":""}'
