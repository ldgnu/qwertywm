#!/bin/sh
set -e

echo "==> qwertywm — Arch Linux install"

echo "==> Instalando dependencias..."
sudo pacman -S --needed go river wayland wayland-protocols libxkbcommon

echo "==> Clonando repo..."
cd /tmp
[ -d qwertywm ] && rm -rf qwertywm
git clone https://github.com/ldgnu/qwertywm.git
cd qwertywm

echo "==> Compilando..."
go build ./cmd/qwertywm
go build ./cmd/qwertywmctl

echo "==> Instalando binarios..."
sudo cp qwertywm qwertywmctl /usr/local/bin/

echo ""
echo "qwertywm instalado. Crea ~/.config/river/init para usarlo."
echo "Ejemplo: https://github.com/ldgnu/qwertywm"
