#!/bin/bash
# qwertywm - instalador one-liner
#   curl -fsSL https://raw.githubusercontent.com/ldgnu/qwertywm/main/install.sh | bash
#
# También podés bajarlo y ejecutarlo local:
#   git clone https://github.com/ldgnu/qwertywm && cd qwertywm && ./install.sh

set -e

# ──────────────────── DETECTAR MODO ────────────────────
# Si estamos en un pipe (curl | bash), el stdin es el script.
# Redirigimos a /dev/tty para poder preguntar al usuario.
if [ ! -f "install.sh" ] || [ "$(basename "$0")" != "install.sh" ] 2>/dev/null; then
    exec </dev/tty
    TMPDIR=$(mktemp -d)
    echo "  Descargando qwertywm..."
    git clone --depth=1 https://github.com/ldgnu/qwertywm.git "$TMPDIR" 2>/dev/null
    cd "$TMPDIR"
    exec bash install.sh
fi

# ──────────────────── SALUDO ────────────────────
clear
cat << "EOF"

   ██████  ██     ██ ███████ ██████  ████████ ██    ██ ██     ██ ██    ██
  ██       ██     ██ ██      ██   ██    ██    ██    ██ ██     ██  ██  ██
  ██   ███ ██  █  ██ █████   ██████     ██     ██  ██  ██  █  ██   ████
  ██    ██ ██ ███ ██ ██      ██   ██    ██      ████   ██ ███ ██    ██
   ██████   ███ ███  ███████ ██   ██    ██       ██     ███ ███  ██    ██

                    qwertywm — Window Manager para River
                Hecho con 💚 por ldgnu (+IA helper)
EOF
echo ""
echo "  Te voy a preguntar un par de cositas y dejo todo listo."
echo "  Si no sabés qué poner, mandale Enter (va la opción default)"
echo ""

# ──────────────────── CHECKEAR ARCH ────────────────────
if ! command -v pacman &>/dev/null; then
    echo "  Este instalador es para Arch Linux."
    echo "  En otras distros, compilá manual:"
    echo "    go build ./cmd/qwertywm"
    echo "    sudo cp qwertywm qwertywmctl /usr/local/bin/"
    echo ""
    exit 1
fi

# ──────────────────── PREGUNTAS ────────────────────
ask() {
    local prompt="$1" default="$2"
    local answer
    read -p "  $prompt [$default]: " answer
    echo "${answer:-$default}"
}

echo "╔════════════════════════════════════╗"
echo "║  1. TERMINAL                       ║"
echo "╚════════════════════════════════════╝"
echo ""
echo "  ¿Qué terminal querés?"
echo "   1) kitty      (bonita, recomendada)"
echo "   2) foot       (la que viene con River)"
echo "   3) alacritty  (rápida, config TOML)"
echo "   4) wezterm    (Rust, config Lua)"
echo "   5) Otra (escribila)"
echo ""
terminal=$(ask "Elegí (1-5)" 1)
case "$terminal" in
    2) terminal="foot" ;;
    3) terminal="alacritty" ;;
    4) terminal="wezterm" ;;
    5) terminal=$(ask "Nombre") ;;
    *) terminal="kitty" ;;
esac

echo ""
echo "╔════════════════════════════════════╗"
echo "║  2. LANZADOR                       ║"
echo "╚════════════════════════════════════╝"
echo ""
echo "  ¿Con qué abrís programas?"
echo "   1) fuzzel   (hecho por el creador de River)"
echo "   2) wofi     (estilo rofi, Wayland)"
echo "   3) bemenu   (minimalista, tipo dmenu)"
echo "   4) No quiero lanzador"
echo ""
launcher=$(ask "Elegí (1-4)" 1)
case "$launcher" in
    2) launcher="wofi" ;;
    3) launcher="bemenu" ;;
    4) launcher="" ;;
    *) launcher="fuzzel" ;;
esac

echo ""
echo "╔════════════════════════════════════╗"
echo "║  3. BARRA                          ║"
echo "╚════════════════════════════════════╝"
echo ""
echo "  ¿Barra con escritorios y hora?"
echo "   1) waybar (clásica, perfil TTY)"
echo "   2) No quiero barra"
echo ""
bar=$(ask "Elegí (1-2)" 1)
[ "$bar" != "2" ] && bar="waybar" || bar=""

echo ""
echo "╔════════════════════════════════════╗"
echo "║  4. COLORES                        ║"
echo "╚════════════════════════════════════╝"
echo ""
echo "  ¿Qué colores te van?"
echo "   1) TTY clásica (negro, verde, gris)"
echo "   2) Catppuccin (pastel)"
echo "   3) Nord (azul frío)"
echo "   4) Solarized (cálido)"
echo ""
theme=$(ask "Elegí (1-4)" 1)
case "$theme" in
    2) theme="catppuccin" ;;
    3) theme="nord" ;;
    4) theme="solarized" ;;
    *) theme="tty" ;;
esac

echo ""
echo "╔════════════════════════════════════╗"
echo "║  5. TECLADO                        ║"
echo "╚════════════════════════════════════╝"
echo ""
echo "  ¿Cómo mover el foco?"
echo "   1) Vim (h=izq, j=abajo, k=arriba, l=der)"
echo "   2) Flechas (←↑↓→)"
echo ""
key=$(ask "Elegí (1-2)" 1)

echo ""
echo "╔════════════════════════════════════╗"
echo "║  6. MONITORES                      ║"
echo "╚════════════════════════════════════╝"
echo ""
multi=$(ask "¿Tenés más de un monitor? (s/N)" n)
if [[ "$multi" =~ ^[sS] ]]; then
    mon1=$(ask "Monitor izquierdo (nombre)" DP-1)
    mon2=$(ask "Monitor derecho (nombre)" HDMI-A-1)
fi

echo ""
echo "╔════════════════════════════════════╗"
echo "║  INSTALANDO...                     ║"
echo "╚════════════════════════════════════╝"
echo ""

# ──────────────────── PAQUETES ────────────────────
packages="river go git wlr-randr ttf-liberation $terminal"
[ -n "$launcher" ] && packages="$packages $launcher"
[ -n "$bar" ] && packages="$packages waybar"

echo "==> Instalando paquetes..."
sudo pacman -S --needed --noconfirm $packages 2>/dev/null

# extras en AUR
if [ "$launcher" = "bemenu" ]; then
    sudo pacman -S --needed --noconfirm bemenu 2>/dev/null || true
fi

# ──────────────────── COMPILAR ────────────────────
echo "==> Compilando..."
go build ./cmd/qwertywm
go build -o qwertywmctl ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/

# ──────────────────── PALETA ────────────────────
case "$theme" in
    tty)
        bg="000000"; fg="c0c0c0"
        bd_f="00aa00"; bd_u="555555"
        match="00aa00"; sel="c0c0c0"; sel_t="000000"
        ;;
    catppuccin)
        bg="1e1e2e"; fg="cdd6f4"
        bd_f="89b4fa"; bd_u="45475a"
        match="89b4fa"; sel="45475a"; sel_t="cdd6f4"
        ;;
    nord)
        bg="2e3440"; fg="d8dee9"
        bd_f="81a1c1"; bd_u="4c566a"
        match="81a1c1"; sel="434c5e"; sel_t="d8dee9"
        ;;
    solarized)
        bg="002b36"; fg="839496"
        bd_f="268bd2"; bd_u="073642"
        match="268bd2"; sel="073642"; sel_t="839496"
        ;;
esac

if [ "$key" = "2" ]; then
    f_l="Left"; f_d="Down"; f_u="Up"; f_r="Right"
else
    f_l="h"; f_d="j"; f_u="k"; f_r="l"
fi

# ──────────────────── CONFIGS ────────────────────
echo "==> Generando configs..."
mkdir -p ~/.config/river ~/.config/qwertywm ~/.config/waybar
[ -n "$launcher" ] && mkdir -p ~/.config/$launcher
mkdir -p ~/.config/foot

# qwertywm config
cat > ~/.config/qwertywm/config << CFG
#!/bin/sh
mod=Super
terminal=$terminal
launcher=${launcher:-none}

qwertywmctl set border-width 2
qwertywmctl set border-color-focused   0x${bd_f}ff
qwertywmctl set border-color-unfocused 0x${bd_u}ff
qwertywmctl set smart-borders on
qwertywmctl set gaps 4 8
qwertywmctl set smart-gaps on
qwertywmctl set main-ratio 0.5
qwertywmctl set main-count 1
qwertywmctl set main-location left
qwertywmctl workspace-mode independent

qwertywmctl bind \$mod+Return  spawn \$terminal
[ -n "$launcher" ] && qwertywmctl bind \$mod+d  spawn '\$launcher'
qwertywmctl bind \$mod+q  close
qwertywmctl bind \$mod+Escape  exit
qwertywmctl bind \$mod+r  spawn 'killall qwertywm 2>/dev/null; qwertywm &; qwertywmctl wait-for-socket; . ~/.config/qwertywm/config &'

qwertywmctl bind \$mod+${f_d}  focus next
qwertywmctl bind \$mod+${f_u}  focus prev
qwertywmctl bind \$mod+${f_l}  focus prev
qwertywmctl bind \$mod+${f_r}  focus next
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

qwertywmctl bind-pointer \$mod+Left  move
qwertywmctl bind-pointer \$mod+Right resize

qwertywmctl bind \$mod+w  focus-output ${mon1:-DP-1}
qwertywmctl bind \$mod+e  focus-output ${mon2:-HDMI-A-1}
CFG
chmod +x ~/.config/qwertywm/config

# bar status script
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

# waybar
if [ -n "$bar" ]; then
cat > ~/.config/waybar/config << WBR
{
    "layer": "bottom",
    "position": "bottom",
    "height": 24,
    "modules-left": ["custom/qwertywm"],
    "modules-right": ["clock", "date"],
    "custom/qwertywm": {
        "exec": "$HOME/.config/qwertywm/bar-status.sh",
        "format": "{}",
        "interval": 1,
        "restart-interval": 5
    },
    "clock": { "interval": 60, "format": "{:%I:%M %p}", "tooltip": false },
    "date": { "interval": 3600, "format": "{:%b %d}", "tooltip": false }
}
WBR

cat > ~/.config/waybar/style.css << CSS
* { font-family: "Liberation Mono"; font-size: 12px; }
window#waybar, #custom-qwertywm, #clock, #date {
    background: #${bg}; color: #${fg}; border: none;
}
#custom-qwertywm, #clock, #date { padding: 0 6px; }
CSS
fi

# fuzzel
if [ "$launcher" = "fuzzel" ]; then
cat > ~/.config/fuzzel/fuzzel.ini << FUZ
[main]
font = Liberation Mono:size=12
prompt = "> "
[colors]
background=${bg}ff; text=${fg}ff; match=${match}ff
selection=${sel}ff; selection-text=${sel_t}ff; border=${bd_f}ff
FUZ
fi

# foot
cat > ~/.config/foot/foot.ini << FOO
[main]
font=Liberation Mono:size=12
[colors]
background=${bg}; foreground=${fg}
regular0=${bd_u}; regular1=aa0000; regular2=00aa00; regular3=aa5500
regular4=0000aa; regular5=aa00aa; regular6=00aaaa; regular7=aaaaaa
bright0=${bd_u}; bright1=ff5555; bright2=55ff55; bright3=ffff55
bright4=5555ff; bright5=ff55ff; bright6=55ffff; bright7=ffffff
selection-foreground=${sel_t}; selection-background=${sel}
FOO

# river init
cat > ~/.config/river/init << RIV
#!/bin/sh
export XDG_CURRENT_DESKTOP=qwertywm
export XDG_SESSION_DESKTOP=qwertywm
${mon1:+wlr-randr --output $mon1 --pos 0,0}
${mon2:+wlr-randr --output $mon2 --pos 1920,0}
${bar:+${bar} &}
qwertywm &
qwertywmctl wait-for-socket
. ~/.config/qwertywm/config
${mon1:+qwertywmctl focus-output $mon1}
qwertywmctl view 1
${mon2:+qwertywmctl focus-output $mon2}
${mon2:+qwertywmctl view 11}
${mon1:+qwertywmctl focus-output $mon1}
RIV
chmod +x ~/.config/river/init

# ──────────────────── DISPLAY MANAGER ────────────────────
echo "==> Instalando sesión display manager..."
sudo tee /usr/share/wayland-sessions/qwertywm.desktop > /dev/null << 'EOF'
[Desktop Entry]
Name=qwertywm
Comment=River Wayland compositor with qwertywm window manager
Exec=env XDG_CURRENT_DESKTOP=qwertywm XDG_SESSION_DESKTOP=qwertywm river
Type=Application
EOF

# ──────────────────── FIN ────────────────────
clear
cat << "EOF"

   ✅  qwertywm instalado con éxito

   Cerra sesión y seleccioná 'qwertywm' en el display manager.
   O desde TTY:   river

   Atajos:
     Super+Enter  → terminal
     Super+d      → lanzador
     Super+j/k    → foco siguiente/anterior
     Super+1..0   → escritorios 1-10
     Alt+1..0     → escritorios 11-20
     Super+r      → recargar config
     Super+Escape → salir

   Editar config: ~/.config/qwertywm/config

EOF
