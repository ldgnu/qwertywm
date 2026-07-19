#!/bin/sh
set -e

echo "==> Instalando qwertywm..."

# Compilar binarios
cd "$(dirname "$0")"
go build ./cmd/qwertywm
go build -o qwertywmctl ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/

# Copiar configs a ~/.config/
echo "==> Copiando configuraciones..."
mkdir -p ~/.config/river
mkdir -p ~/.config/qwertywm
mkdir -p ~/.config/waybar
mkdir -p ~/.config/fuzzel
mkdir -p ~/.config/foot

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

# Sesión para display manager
echo "==> Instalando sesión para display manager..."
sudo tee /usr/share/wayland-sessions/qwertywm.desktop > /dev/null << 'EOF'
[Desktop Entry]
Name=qwertywm
Comment=River compositor with qwertywm window manager
Exec=river
Type=Application
EOF

echo ""
echo "======================"
echo "qwertywm instalado!"
echo "======================"
echo ""
echo "Usar:"
echo "  1. Cerrar sesión actual"
echo "  2. Seleccionar 'qwertywm' en el display manager"
echo ""
echo "O desde TTY:"
echo "  XDG_RUNTIME_DIR=/run/user/\$(id -u) river"
echo ""
echo "Atajos:"
echo "  Super+Enter  = terminal (kitty)"
echo "  Super+d      = lanzador (fuzzel)"
echo "  Super+j/k    = foco siguiente/anterior"
echo "  Super+h/l    = mover foco izq/der"
echo "  Super+1..0   = escritorios 1-10"
echo "  Alt+1..0     = escritorios 11-20"
echo "  Super+Space  = cambiar layout"
echo "  Super+f      = fullscreen"
echo "  Super+q      = cerrar ventana"
echo "  Super+r      = recargar config"
echo "  Super+Escape = salir"
echo "  Super+Click  = mover/redimensionar"
echo ""
echo "Requisitos: river, kitty, fuzzel, waybar, foot,"
echo "            wlr-randr, Liberation Mono"
