#!/bin/sh
set -e

echo "==> qwertywm build"
echo ""

build() {
    echo "  go build ./cmd/$1"
    go build -ldflags="-s -w" ./cmd/$1
}

build qwertywm
build qwertywmctl

echo ""
echo "Listo:"
ls -lh qwertywm qwertywmctl
echo ""
echo "Instalar: sudo cp qwertywm qwertywmctl /usr/local/bin/"
