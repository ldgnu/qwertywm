#!/bin/sh
data=$(qwertywmctl get state 2>/dev/null)
foc_ws=$(echo "$data" | jq -r '[.outputs[] | select(.focused) | .workspace][0]' 2>/dev/null)
title=$(echo "$data" | jq -r '[.windows[] | select(.focused) | .title][0]' 2>/dev/null)

tags=$(echo "$data" | jq -r '
  [.workspaces[] | select(.windows != null or .visible) | .name] | sort_by(tonumber) | .[]
' 2>/dev/null | while read ws; do
  if [ "$ws" = "$foc_ws" ]; then echo "[$ws]"; else echo "$ws"; fi
done | tr '\n' ' ')

echo "{\"text\":\"$tags $title\"}"
