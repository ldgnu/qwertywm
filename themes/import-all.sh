#!/bin/sh
JSON="/home/ldgnu/.local/share/opencode/tool-output/tool_f164b33ec001GdMU3wCic9tjVO"
OUTDIR="$HOME/.config/themes"

jq -c '.[]' "$JSON" | while read -r t; do
  name=$(echo "$t" | jq -r '.name' | tr ' ' '_' | tr -d '()/,')
  [ -z "$name" ] && continue

  bg=$(echo "$t" | jq -r '.background // "#000000"')
  fg=$(echo "$t" | jq -r '.foreground // "#c0c0c0"')
  cur=$(echo "$t" | jq -r '.cursor // "#c0c0c0"')

  cat > "$OUTDIR/${name}.conf" << EOF
name=$(echo "$t" | jq -r '.name')
background=$bg
foreground=$fg
cursor=$cur
EOF

  for i in $(seq 1 16); do
    c=$(echo "$t" | jq -r ".color_$(printf '%02d' $i) // \"\"" 2>/dev/null)
    [ -z "$c" ] || [ "$c" = "null" ] && c="$fg"
    echo "color$((i - 1))=$c" >> "$OUTDIR/${name}.conf"
  done
done

echo "Done: $(ls "$OUTDIR"/*.conf | wc -l) themes"
