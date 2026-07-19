#!/bin/sh
data=$(qwertywmctl get state 2>/dev/null)
w1=$(echo "$data" | jq -r '.outputs[0].workspace' 2>/dev/null)
w2=$(echo "$data" | jq -r '.outputs[1].workspace' 2>/dev/null)
foc=$(echo "$data" | jq -r '[.outputs[] | select(.focused) | .workspace][0]' 2>/dev/null)
title=$(echo "$data" | jq -r '[.windows[] | select(.focused) | .title][0]' 2>/dev/null)

if [ "$foc" = "$w1" ]; then
  echo "{\"text\":\"[$w1] $w2  $title\"}"
else
  echo "{\"text\":\"$w1 [$w2]  $title\"}"
fi
