#!/bin/bash
set -e

echo "╔══════════════════════════════════════╗"
echo "║       qwertywm - Instalador          ║"
echo "╚══════════════════════════════════════╝"
echo ""

# Detectar si estamos en Arch
if ! command -v pacman &>/dev/null; then
    echo "Este instalador es para Arch Linux."
    echo "En otras distros, compile manualmente:"
    echo "  go build ./cmd/qwertywm"
    echo "  go build -o qwertywmctl ./cmd/qwertywmctl"
    echo "  sudo cp qwertywm qwertywmctl /usr/local/bin/"
    exit 1
fi

# ──────────────────── 1. INSTALAR DEPENDENCIAS ────────────────────
echo "==> Instalando dependencias..."
sudo pacman -S --needed --noconfirm \
    river \
    kitty \
    fuzzel \
    foot \
    waybar \
    wlr-randr \
    ttf-liberation \
    go \
    git \
    base-devel

# ──────────────────── 2. COMPILAR ────────────────────
echo "==> Compilando qwertywm..."
cd "$(dirname "$0")"
go build ./cmd/qwertywm
go build -o qwertywmctl ./cmd/qwertywmctl

# ──────────────────── 3. INSTALAR BINARIOS ────────────────────
echo "==> Instalando binarios..."
sudo cp qwertywm qwertywmctl /usr/local/bin/

# ──────────────────── 4. CONFIGS ────────────────────
echo "==> Copiando configuraciones a ~/.config/..."
mkdir -p ~/.config/river
mkdir -p ~/.config/qwertywm
mkdir -p ~/.config/waybar
mkdir -p ~/.config/fuzzel
mkdir -p ~/.config/foot
mkdir -p ~/.config/kitty

cp config/river/init ~/.config/river/init
cp config/qwertywm/config ~/.config/qwertywm/config
cp config/qwertywm/bar-status.sh ~/.config/qwertywm/bar-status.sh
cp config/waybar/config ~/.config/waybar/config
cp config/waybar/style.css ~/.config/waybar/style.css
cp config/fuzzel/fuzzel.ini ~/.config/fuzzel/fuzzel.ini
cp config/foot/foot.ini ~/.config/foot/foot.ini

chmod +x ~/.config/river/init
chmod +x ~/.config/qwertywm/config
chmod +x ~/.config/qwertywm/bar-status.sh

# ──────────────────── 5. SESIÓN DISPLAY MANAGER ────────────────────
echo "==> Instalando entrada para display manager..."
sudo tee /usr/share/wayland-sessions/qwertywm.desktop > /dev/null << 'EOF'
[Desktop Entry]
Name=qwertywm
Comment=River Wayland compositor with qwertywm window manager
Exec=river
Type=Application
EOF

# ──────────────────── 6. VERIFICACIÓN ────────────────────
echo ""
echo "╔══════════════════════════════════════╗"
echo "║  qwertywm instalado correctamente   ║"
echo "╚══════════════════════════════════════╝"
echo ""
echo "  Para USAR qwertywm:"
echo ""
echo "  A) Desde el display manager (GDM/SDDM/LightDM):"
echo "     - Cerrar sesión"
echo "     - Seleccionar 'qwertywm' del menú de sesiones"
echo "     - Iniciar sesión"
echo ""
echo "  B) Desde TTY (Ctrl+Alt+F2):"
echo "     $ startx"
echo "     O directamente:"
echo "     $ XDG_RUNTIME_DIR=/run/user/\$(id -u) river"
echo ""
echo "  Una vez dentro:"
echo ""
echo "  ┌──────────────┬──────────────────────────┐"
echo "  │  Atajo        │  Acción                  │"
echo "  ├──────────────┼──────────────────────────┤"
echo "  │ Super+Enter  │  Terminal (kitty)         │"
echo "  │ Super+d      │  Lanzador (fuzzel)        │"
echo "  │ Super+j/k    │  Foco siguiente/anterior  │"
echo "  │ Super+h/l    │  Mover foco izq/der       │"
echo "  │ Super+w/e    │  Focus DP-1 / HDMI        │"
echo "  │ Super+1..0   │  Escritorios 1-10         │"
echo "  │ Alt+1..0     │  Escritorios 11-20        │"
echo "  │ Super+Space  │  Cambiar layout           │"
echo "  │ Super+f      │  Fullscreen               │"
echo "  │ Super+q      │  Cerrar ventana            │"
echo "  │ Super+r      │  Recargar configuración    │"
echo "  │ Super+Escape │  Salir del WM             │"
echo "  │ Super+Click  │  Mover/redimensionar      │"
echo "  └──────────────┴──────────────────────────┘"
echo ""
echo "  Editar config: ~/.config/qwertywm/config"
echo "  Recargar: Super+r"
echo ""
echo "  Más info: https://github.com/ldgnu/qwertywm"
