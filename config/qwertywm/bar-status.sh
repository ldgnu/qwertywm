#!/bin/sh
data=$(qwertywmctl get state 2>/dev/null)
tags=$(echo "$data" | jq -r '[.outputs[] | if .focused then "[" + .workspace + "]" else .workspace end] | join(" ")' 2>/dev/null)
title=$(echo "$data" | jq -r '[.windows[] | select(.focused) | .title][0]' 2>/dev/null)
echo '{"text":"'"$tags  $title"'"}'
