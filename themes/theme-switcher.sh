#!/bin/sh
THEME_DIR="$HOME/.config/themes"

theme=$(ls "$THEME_DIR"/*.conf | while read f; do
  name=$(grep '^name=' "$f" | cut -d= -f2)
  bg=$(grep '^background=' "$f" | cut -d= -f2)
  fg=$(grep '^foreground=' "$f" | cut -d= -f2)
  echo "$name  $bg $fg"
done | fuzzel --dmenu --prompt="theme: " --width=60 --lines=10 | awk '{print $1}')

[ -z "$theme" ] && exit 0

"$THEME_DIR/apply-theme.sh" "$theme"

# Notify
notify-send -t 2000 "Theme: $theme" "Applied to kitty"
