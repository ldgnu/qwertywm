#!/bin/sh
# Apply a Gogh theme to kitty, waybar, Zen, and Telegram
THEME_DIR="$HOME/.config/themes"
TPL_DIR="$THEME_DIR/templates"
KITTY_CONF="$HOME/.config/kitty/kitty.conf"
WAYBAR_CSS="$HOME/.config/waybar/style.css"
ZEN_CHROME="$HOME/.config/zen/tvntjdhm.Default (release)/chrome/userChrome.css"
TELEGRAM_THEME_DIR="$HOME/.config/telegram-tty-theme"

if [ -z "$1" ]; then echo "Usage: $0 <theme-name>"; exit 1; fi

THEME_FILE="$THEME_DIR/$1.conf"
if [ ! -f "$THEME_FILE" ]; then
  echo "Theme '$1' not found"; ls "$THEME_DIR"/*.conf | sed 's/.*\///;s/\.conf$//'; exit 1
fi

. "$THEME_FILE"

# Helper: lighten a hex color by mixing with white
lighten() {
  local hex=$1; hex=${hex#\#}
  local r=$((16#${hex:0:2})) g=$((16#${hex:2:2})) b=$((16#${hex:4:2}))
  r=$((r + (255 - r) * 30 / 100))
  g=$((g + (255 - g) * 30 / 100))
  b=$((b + (255 - b) * 30 / 100))
  printf "#%02x%02x%02x" $r $g $b
}

BLACK="#000000"
GRAY=$(lighten "$background")
[ "$GRAY" = "$background" ] && GRAY="#333333"
BG="$background"; FG="$foreground"
GREEN="${color2:-$foreground}"
CYAN="${color6:-$foreground}"
BLUE="${color4:-$foreground}"
YELLOW="${color3:-$foreground}"
RED="${color1:-$foreground}"
MAGENTA="${color5:-$foreground}"
BG_LIGHT=$(lighten "$BG")

# === KITTY ===
sed -i '/^# === THEME COLORS START ===/,/^# === THEME COLORS END ===/d' "$KITTY_CONF"
cat >> "$KITTY_CONF" << KITTYEOF

# === THEME COLORS START ===
# Current theme: $name
foreground $FG
background $BG
cursor $FG
color0 $color0
color1 $color1
color2 $color2
color3 $color3
color4 $color4
color5 $color5
color6 $color6
color7 $color7
color8 $color8
color9 $color9
color10 $color10
color11 $color11
color12 $color12
color13 $color13
color14 $color14
color15 $color15
# === THEME COLORS END ===
KITTYEOF

# === WAYBAR ===
sed -e "s/\$BG/$BG/g" -e "s/\$FG/$FG/g" \
    -e "s/\$GREEN/$GREEN/g" -e "s/\$CYAN/$CYAN/g" \
    -e "s/\$BLUE/$BLUE/g" -e "s/\$YELLOW/$YELLOW/g" \
    -e "s/\$RED/$RED/g" -e "s/\$MAGENTA/$MAGENTA/g" \
    "$TPL_DIR/waybar.css.tpl" > "$WAYBAR_CSS"

# === ZEN BROWSER ===
sed -e "s/\$BG/$BG/g" -e "s/\$FG/$FG/g" \
    -e "s/\$GREEN/$GREEN/g" -e "s/\$CYAN/$CYAN/g" \
    -e "s/\$BLUE/$BLUE/g" -e "s/\$YELLOW/$YELLOW/g" \
    -e "s/\$RED/$RED/g" -e "s/\$BG_LIGHT/$BG_LIGHT/g" \
    "$TPL_DIR/zen.css.tpl" > "$ZEN_CHROME"

# === TELEGRAM ===
mkdir -p "$TELEGRAM_THEME_DIR"
sed -e "s/\$BG/$BG/g" -e "s/\$FG/$FG/g" \
    -e "s/\$GREEN/$GREEN/g" -e "s/\$CYAN/$CYAN/g" \
    -e "s/\$BLUE/$BLUE/g" -e "s/\$YELLOW/$YELLOW/g" \
    -e "s/\$RED/$RED/g" -e "s/\$BLACK/$BLACK/g" \
    -e "s/\$GRAY/$GRAY/g" -e "s/\$BG_LIGHT/$BG_LIGHT/g" \
    -e "s/\$MAGENTA/$MAGENTA/g" \
    "$TPL_DIR/telegram.colors.tpl" > "$TELEGRAM_THEME_DIR/colors.tdesktop"
cd "$TELEGRAM_THEME_DIR" && zip -j telegram-tty-theme.tdesktop-theme colors.tdesktop 2>/dev/null

# === QTEBROWSER ===
sed -e "s/\$BG/$BG/g" -e "s/\$FG/$FG/g" \
    -e "s/\$GREEN/$GREEN/g" -e "s/\$CYAN/$CYAN/g" \
    -e "s/\$BLUE/$BLUE/g" -e "s/\$YELLOW/$YELLOW/g" \
    -e "s/\$RED/$RED/g" -e "s/\$BG_LIGHT/$BG_LIGHT/g" \
    "$TPL_DIR/qutebrowser.py.tpl" > "$HOME/.config/qutebrowser/config.py"

# === GTK (Thunar, etc) ===
sed -e "s/\$BG/$BG/g" -e "s/\$FG/$FG/g" \
    -e "s/\$GREEN/$GREEN/g" -e "s/\$CYAN/$CYAN/g" \
    -e "s/\$BLUE/$BLUE/g" -e "s/\$YELLOW/$YELLOW/g" \
    -e "s/\$RED/$RED/g" -e "s/\$BG_LIGHT/$BG_LIGHT/g" \
    "$TPL_DIR/gtk.css.tpl" > "$HOME/.config/gtk-3.0/gtk.css"

# === FUZZEL ===
sed -e "s/\${BG}/$BG/g" -e "s/\${FG}/$FG/g" -e "s/\${GREEN}/$GREEN/g" \
    "$TPL_DIR/fuzzel.ini.tpl" > "$HOME/.config/fuzzel/fuzzel.ini"

# === QWERTYWM ===
qwertywmctl set border-color-focused "$GREEN"
qwertywmctl set border-color-unfocused "$GRAY"

# === RESTART WAYBAR ===
killall waybar 2>/dev/null; nohup waybar > /dev/null 2>&1 &

echo "Theme '$name' applied to kitty + waybar + Zen + Telegram"
echo "Restart kitty windows. Import .config/telegram-tty-theme/telegram-tty-theme.tdesktop-theme in Telegram."
