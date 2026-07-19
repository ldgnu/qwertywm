#!/bin/bash
# qwertywm - instalador interactivo
# "Hice que una IA lo escribiera, pero la onda es toda mía" ©

set -e
cd "$(dirname "$0")"

# ============================
#  SALUDO
# ============================
clear
echo "╔═══════════════════════════════════════════════╗"
echo "║                                               ║"
echo "║     ██████  ██     ██ ███████ ██████          ║"
echo "║    ██       ██     ██ ██      ██   ██         ║"
echo "║    ██   ███ ██  █  ██ █████   ██████          ║"
echo "║    ██    ██ ██ ███ ██ ██      ██   ██         ║"
echo "║     ██████   ███ ███  ███████ ██   ██         ║"
echo "║                                               ║"
echo "║     qwertywm - Window Manager para River      ║"
echo "║                                               ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  Hola 👋 Este script te va a preguntar cositas"
echo "  y deja todo listo para que arranques."
echo ""
echo "  Si en cualquier momento no sabés qué elegir,"
echo "  mandale Enter nomas (la opción por defecto)"
echo ""
read -p "  [Enter] para arrancar → "

# ============================
#  DETECTAR SISTEMA
# ============================
if ! command -v pacman &>/dev/null; then
    echo ""
    echo "  Uy, esto no parece Arch Linux 🤔"
    echo "  El instalador automático anda solo en Arch,"
    echo "  pero podés compilarlo a mano:"
    echo "    go build ./cmd/qwertywm"
    echo "    sudo cp qwertywm qwertywmctl /usr/local/bin/"
    echo ""
    exit 1
fi

# ============================
#  PREGUNTAS
# ============================
echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║           1. TERMINAL                         ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  ¿Qué terminal querés usar?"
echo "  1) kitty  (el más lindo, recomendado)"
echo "  2) foot   (el que viene con River, más ligero)"
echo "  3) alacritty (rápido, config en TOML)"
echo "  4) wezterm (hecho en Rust, config en Lua)"
echo "  5) Otra (la pongo como sea que se llame)"
echo ""
read -p "  [1]: " terminal_choice
case "$terminal_choice" in
    2) terminal="foot" ;;
    3) terminal="alacritty" ;;
    4) terminal="wezterm" ;;
    5) read -p "  Decime el nombre: " terminal ;;
    *) terminal="kitty" ;;
esac

echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║           2. LANZADOR DE APPS                 ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  ¿Con qué querés abrir programas?"
echo "  1) fuzzel   (hecho por el mismo de River)"
echo "  2) wofi     (estilo rofi pero Wayland)"
echo "  3) bemenu   (minimalista, tipo dmenu)"
echo "  4) tofi     (otro dmenu-like, bonito)"
echo "  5) No quiero lanzador (?)"
echo ""
read -p "  [1]: " launch_choice
case "$launch_choice" in
    2) launcher="wofi" ;;
    3) launcher="bemenu" ;;
    4) launcher="tofi" ;;
    5) launcher="none" ;;
    *) launcher="fuzzel" ;;
esac

echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║           3. BARRA DE ESTADO                  ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  ¿Barra abajo con escritorios y hora?"
echo "  1) waybar (la clásica, perfil TTY)"
echo "  2) eww    (si te gusta sufrir configurando)"
echo "  3) No quiero barra (sobrio)"
echo ""
read -p "  [1]: " bar_choice
case "$bar_choice" in
    2) bar="eww" ;;
    3) bar="none" ;;
    *) bar="waybar" ;;
esac

echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║           4. COLORES                          ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  ¿Qué onda de colores?"
echo "  1) TTY clásica (fondo negro, verde terminal)"
echo "  2) Catppuccin (pastel, modo彻)"
echo "  3) Nord (azulitos fríos)"
echo "  4) Solarized (amarillito, para leer)"
echo ""
read -p "  [1]: " color_choice
case "$color_choice" in
    2) theme="catppuccin" ;;
    3) theme="nord" ;;
    4) theme="solarized" ;;
    *) theme="tty" ;;
esac

echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║           5. TECLADO                          ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  ¿Cómo mover el foco entre ventanas?"
echo "  1) Vim (h=izq, j=abajo, k=arriba, l=der)"
echo "  2) Flechas (←↑↓→)"
echo ""
read -p "  [1]: " key_choice

echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║           6. MONITORES                        ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  ¿Tenés más de un monitor?"
read -p "  [s/N]: " multi_mon
if [[ "$multi_mon" =~ ^[sSyY] ]]; then
    echo ""
    echo "  Buenísimo. ¿Cuál es el nombre del monitor"
    echo "  principal (izquierda)?"
    echo "  Tips: después podés verlos con 'wlr-randr'"
    echo "  Dejalo vacío si querés configurar después"
    read -p "  [DP-1]: " mon_primary
    [ -z "$mon_primary" ] && mon_primary="DP-1"
    echo "  ¿Y el secundario (derecha)?"
    read -p "  [HDMI-A-1]: " mon_secondary
    [ -z "$mon_secondary" ] && mon_secondary="HDMI-A-1"
fi

# ============================
#  RESUMEN
# ============================
echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║           7. TODO LISTO                       ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  Resumen:"
echo "  • Terminal:    $terminal"
echo "  • Lanzador:    $launcher"
echo "  • Barra:       $bar"
echo "  • Colores:     $theme"
echo "  • Teclas:      $([ "$key_choice" = "2" ] && echo "Flechas" || echo "Vim")"
echo "  • Monitores:   ${mon_primary:-ninguno} / ${mon_secondary:-ninguno}"
echo ""
read -p "  ¿Instalamos? [Enter] → "

# ============================
#  INSTALAR DEPENDENCIAS
# ============================
echo ""
echo "==> Instalando paquetes..."
packages="river go git wlr-randr ttf-liberation $terminal"
[ "$launcher" != "none" ] && packages="$packages $launcher"
[ "$bar" = "waybar" ] && packages="$packages waybar"

# Launchers extra que quizás no están en pacman
if [ "$launcher" = "tofi" ]; then
    packages="$packages tofi"  # está en community
fi

sudo pacman -S --needed --noconfirm $packages

# bemenu está en repos
if [ "$launcher" = "bemenu" ]; then
    sudo pacman -S --needed --noconfirm bemenu
fi

# eww no está en pacman, se saltea
if [ "$bar" = "eww" ]; then
    echo "  EWW no está en pacman. Instalalo a mano."
    echo "  Por ahora te pongo waybar."
    bar="waybar"
    sudo pacman -S --needed --noconfirm waybar
fi

# ============================
#  COMPILAR
# ============================
echo ""
echo "==> Compilando qwertywm..."
go build ./cmd/qwertywm
go build -o qwertywmctl ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/

# ============================
#  GENERAR CONFIGS
# ============================
echo ""
echo "==> Generando configuraciones..."

mkdir -p ~/.config/river
mkdir -p ~/.config/qwertywm
mkdir -p ~/.config/waybar
mkdir -p ~/.config/$terminal
[ "$launcher" != "none" ] && mkdir -p ~/.config/$launcher

# ---- COLORES ----
# TTY classic
if [ "$theme" = "tty" ]; then
    bg="000000"
    fg="c0c0c0"
    bd_focused="00aa00"
    bd_unfocused="555555"
    match="00aa00"
    selection="c0c0c0"
    sel_text="000000"
# Catppuccin Mocha
elif [ "$theme" = "catppuccin" ]; then
    bg="1e1e2e"
    fg="cdd6f4"
    bd_focused="89b4fa"
    bd_unfocused="45475a"
    match="89b4fa"
    selection="45475a"
    sel_text="cdd6f4"
# Nord
elif [ "$theme" = "nord" ]; then
    bg="2e3440"
    fg="d8dee9"
    bd_focused="81a1c1"
    bd_unfocused="4c566a"
    match="81a1c1"
    selection="434c5e"
    sel_text="d8dee9"
# Solarized
elif [ "$theme" = "solarized" ]; then
    bg="002b36"
    fg="839496"
    bd_focused="268bd2"
    bd_unfocused="073642"
    match="268bd2"
    selection="073642"
    sel_text="839496"
fi

# Si es vim o flechas para los bindings de enfoque
if [ "$key_choice" = "2" ]; then
    focus_left="Left"
    focus_down="Down"
    focus_up="Up"
    focus_right="Right"
else
    focus_left="h"
    focus_down="j"
    focus_up="k"
    focus_right="l"
fi

# Tecla Super
mod="Super"

# ---- QWERTYWM CONFIG ----
cat > ~/.config/qwertywm/config << CFG
#!/bin/sh
# qwertywm config - generado por install.sh
# Editá esto y reiniciá con Super+r

mod=$mod
terminal=$terminal
launcher=$launcher

qwertywmctl set border-width 2
qwertywmctl set border-color-focused   0x${bd_focused}ff
qwertywmctl set border-color-unfocused 0x${bd_unfocused}ff
qwertywmctl set smart-borders on
qwertywmctl set gaps 4 8
qwertywmctl set smart-gaps on
qwertywmctl set main-ratio 0.5
qwertywmctl set main-count 1
qwertywmctl set main-location left
qwertywmctl workspace-mode independent

qwertywmctl bind \$mod+Return          spawn \$terminal
qwertywmctl bind \$mod+d               spawn '\$launcher'
qwertywmctl bind \$mod+q               close
qwertywmctl bind \$mod+Escape          exit
qwertywmctl bind \$mod+r               spawn 'killall qwertywm 2>/dev/null; qwertywm &; qwertywmctl wait-for-socket; . ~/.config/qwertywm/config &'

qwertywmctl bind \$mod+$focus_down  focus next
qwertywmctl bind \$mod+$focus_up    focus prev
qwertywmctl bind \$mod+$focus_left  focus prev
qwertywmctl bind \$mod+$focus_right focus next
qwertywmctl bind \$mod+space  cycle-layout monocle,left,top
qwertywmctl bind \$mod+f  toggle-fullscreen

for i in 1 2 3 4 5 6 7 8 9; do
    qwertywmctl bind \$mod+\$i view \$i
    qwertywmctl bind \$mod+Shift+\$i send \$i
done
qwertywmctl bind \$mod+0  view 10
qwertywmctl bind \$mod+Shift+0  send 10

for i in 1 2 3 4 5 6 7 8 9; do
    n=\$((i + 10))
    qwertywmctl bind Alt+\$i view \$n
    qwertywmctl bind Alt+Shift+\$i send \$n
done
qwertywmctl bind Alt+0  view 20
qwertywmctl bind Alt+Shift+0  send 20

qwertywmctl bind-pointer \$mod+Left   move
qwertywmctl bind-pointer \$mod+Right  resize

qwertywmctl bind \$mod+w  focus-output ${mon_primary:-DP-1}
qwertywmctl bind \$mod+e  focus-output ${mon_secondary:-HDMI-A-1}
qwertywmctl bind \$mod+Shift+w  send-to-output ${mon_primary:-DP-1}
qwertywmctl bind \$mod+Shift+e  send-to-output ${mon_secondary:-HDMI-A-1}
CFG
chmod +x ~/.config/qwertywm/config

# ---- BAR SCRIPT ----
cat > ~/.config/qwertywm/bar-status.sh << 'BAR'
#!/bin/sh
qwertywmctl subscribe | while read -r line; do
  echo "$line" | jq -c '
    . as $d |
    [ $d.outputs[] | "\(.workspace)" ] | join(" ") as $tags |
    ( $d.windows[] | select(.focused) | .title ) as $title |
    { text: "\($tags)  \($title)" }
  ' 2>/dev/null || echo '{"text":""}'
done
BAR
chmod +x ~/.config/qwertywm/bar-status.sh

# ---- WAYBAR CONFIG ----
if [ "$bar" = "waybar" ]; then
cat > ~/.config/waybar/config << WBR
{
    "layer": "bottom",
    "position": "bottom",
    "height": 24,
    "modules-left": ["custom/qwertywm"],
    "modules-center": [],
    "modules-right": ["clock", "date"],
    "custom/qwertywm": {
        "exec": "$HOME/.config/qwertywm/bar-status.sh",
        "format": "{}",
        "interval": 1,
        "restart-interval": 5
    },
    "clock": {
        "interval": 60,
        "format": "{:%I:%M %p}",
        "tooltip": false
    },
    "date": {
        "interval": 3600,
        "format": "{:%b %d}",
        "tooltip": false
    }
}
WBR

cat > ~/.config/waybar/style.css << CSS
* {
    font-family: "Liberation Mono";
    font-size: 12px;
}
window#waybar {
    background: #${bg};
    color: #${fg};
    border: none;
}
#custom-qwertywm {
    color: #${fg};
    background: #${bg};
    padding: 0 6px;
}
#custom-qwertywm .tag-active {
    color: #${bd_focused};
    font-weight: bold;
}
#clock, #date {
    color: #${fg};
    background: #${bg};
    padding: 0 6px;
}
CSS
fi

# ---- FUZZEL CONFIG ----
if [ "$launcher" = "fuzzel" ]; then
cat > ~/.config/fuzzel/fuzzel.ini << FUZ
[main]
font = Liberation Mono:size=12
prompt = "> "

[colors]
background=${bg}ff
text=${fg}ff
match=${bd_focused}ff
selection=${selection}ff
selection-text=${sel_text}ff
border=${bd_focused}ff
FUZ
fi

# ---- FOOT CONFIG (de paso) ----
cat > ~/.config/foot/foot.ini << FOO
[main]
font=Liberation Mono:size=12

[colors]
background=${bg}
foreground=${fg}
regular0=${bd_unfocused}
regular1=aa0000
regular2=00aa00
regular3=aa5500
regular4=0000aa
regular5=aa00aa
regular6=00aaaa
regular7=aaaaaa
bright0=${bd_unfocused}
bright1=ff5555
bright2=55ff55
bright3=ffff55
bright4=5555ff
bright5=ff55ff
bright6=55ffff
bright7=ffffff
selection-foreground=${sel_text}
selection-background=${selection}
FOO

# ---- RIVER INIT ----
cat > ~/.config/river/init << RIV
#!/bin/sh
export XDG_CURRENT_DESKTOP=qwertywm
export XDG_SESSION_DESKTOP=qwertywm

\${mon_primary:+wlr-randr --output \$mon_primary --pos 0,0 --mode 1920x1080}
\${mon_secondary:+wlr-randr --output \$mon_secondary --pos 1920,0 --mode 1920x1080}
\${bar:+\${bar} &}
qwertywm &
qwertywmctl wait-for-socket
. ~/.config/qwertywm/config

${mon_primary:+qwertywmctl focus-output $mon_primary}
qwertywmctl view 1
${mon_secondary:+qwertywmctl focus-output $mon_secondary}
${mon_secondary:+qwertywmctl view 11}
${mon_primary:+qwertywmctl focus-output $mon_primary}
RIV
chmod +x ~/.config/river/init

# ---- SESIÓN DISPLAY MANAGER ----
echo "==> Instalando sesión para display manager..."
sudo tee /usr/share/wayland-sessions/qwertywm.desktop > /dev/null << 'EOF'
[Desktop Entry]
Name=qwertywm
Comment=River compositor with qwertywm window manager
Exec=env XDG_CURRENT_DESKTOP=qwertywm XDG_SESSION_DESKTOP=qwertywm river
Type=Application
EOF

# ============================
#  FIN
# ============================
clear
echo "╔═══════════════════════════════════════════════╗"
echo "║                                               ║"
echo "║   ✅  qwertywm instalado con éxito           ║"
echo "║                                               ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo "  ¿Qué sigue?"
echo ""
echo "  1. Cerrá la sesión actual"
echo "  2. En el display manager, seleccioná 'qwertywm'"
echo "     (Si no aparece, reiniciá)"
echo ""
echo "  O desde una TTY (Ctrl+Alt+F2):"
echo "    river"
echo ""
echo "  Atajos que te quedaron:"
echo "    Super+Enter  → $terminal"
echo "    Super+d      → $launcher"
echo "    Super+j/k    → foco siguiente/anterior"
echo "    Super+1..0   → escritorios 1-10"
echo "    Alt+1..0     → escritorios 11-20"
echo "    Super+r      → recargar config"
echo "    Super+Escape → salir"
echo ""
echo "  Editar config: ~/.config/qwertywm/config"
echo ""
echo "  Hecho con 💚  por una IA, pero la idea es mía"
