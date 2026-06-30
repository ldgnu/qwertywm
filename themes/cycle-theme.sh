#!/bin/sh
# Cycle through themes: next or prev
DIR="$HOME/.config/themes"
CACHE="$DIR/.current"
STEP="${1:-next}"

[ ! -f "$CACHE" ] && echo 0 > "$CACHE"
idx=$(cat "$CACHE")
themes=$(ls "$DIR"/*.conf | sort)
count=$(echo "$themes" | wc -l)

if [ "$STEP" = "next" ]; then
  idx=$(( (idx + 1) % count ))
else
  idx=$(( (idx - 1 + count) % count ))
fi

echo $idx > "$CACHE"
theme=$(echo "$themes" | sed -n "$((idx + 1))p" | xargs basename -s .conf)
"$DIR/apply-theme.sh" "$theme"
