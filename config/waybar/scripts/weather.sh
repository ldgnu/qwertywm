#!/bin/sh
data=$(curl -s --max-time 3 "wttr.in/Cordoba,Argentina?format=%t+%C" 2>/dev/null)
[ -z "$data" ] && data="--"
echo "{\"text\":\"${data}\"}"
