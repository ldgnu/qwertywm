#!/bin/sh
data=$(qwertywmctl get state 2>/dev/null)

hname=$(echo "$data" | jq -r '.outputs[0].name' 2>/dev/null)
hw=$(echo "$data" | jq -r '.outputs[0].workspace' 2>/dev/null)
hfoc=$(echo "$data" | jq -r '.outputs[0].focused' 2>/dev/null)

dname=$(echo "$data" | jq -r '.outputs[1].name' 2>/dev/null)
dw=$(echo "$data" | jq -r '.outputs[1].workspace' 2>/dev/null)
dfoc=$(echo "$data" | jq -r '.outputs[1].focused' 2>/dev/null)

title=$(echo "$data" | jq -r '[.windows[] | select(.focused) | .title][0]' 2>/dev/null)

# HDMI (izquierda) = verde, DP (derecha) = azul
# El monitor enfocado muestra corchetes
if [ "$hfoc" = "true" ]; then
  hdmi="[$hw]"
else
  hdmi="$hw"
fi
if [ "$dfoc" = "true" ]; then
  dp="[$dw]"
else
  dp="$dw"
fi

echo "{\"text\":\"$hdmi  $dp  $title\"}"
