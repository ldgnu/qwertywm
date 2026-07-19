#!/bin/sh
data=$(qwertywmctl get state 2>/dev/null)

ws=$(echo "$data" | jq -r '.outputs[] | select(.focused) | .workspace' 2>/dev/null)
title=$(echo "$data" | jq -r '[.windows[] | select(.focused) | .title][0]' 2>/dev/null)

echo "{\"text\":\"[$ws]  $title\"}"
