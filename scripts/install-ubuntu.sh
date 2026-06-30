#!/bin/sh
set -e

echo "==> qwertywm — Ubuntu 24.04 install"

echo "==> Instalando dependencias..."
sudo apt update
sudo apt install -y golang-go river wayland-protocols libxkbcommon-dev

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
