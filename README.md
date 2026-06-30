# qwertywm

Un window manager dinámico estilo xmonad para [river](https://codeberg.org/river/river),
escrito en Go. Tiling dinámico, configuración programática vía socket unix,
soporte multi-monitor de primera clase.

## Créditos

**qwertywm** es un fork de [weir](https://github.com/psanford/weir) por
[psanford](https://github.com/psanford). El proyecto original fue renombrado
y modificado para uso personal por [ldgnu](https://github.com/ldgnu).

Gracias a psanford por el laburazo del core, la arquitectura limpia y toda la
base sólida. Este fork solo cambia nombres, afina detalles y adapta el setup.

---

## Requisitos

- [river](https://codeberg.org/river/river) ≥ 0.4
- Go ≥ 1.21 (solo para compilar)
- `wayland` (protocolos)
- `wayland-protocols`
- `libxkbcommon` (para `keyboard-layout`)

## Instalación

### Arch Linux / CachyOS

```sh
# Dependencias
sudo pacman -S go river wayland wayland-protocols libxkbcommon

# Compilar e instalar
git clone https://github.com/ldgnu/qwertywm.git
cd qwertywm
go build ./cmd/qwertywm
go build ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/

# Config
mkdir -p ~/.config/river
cat > ~/.config/river/init << 'EOF'
#!/bin/sh
wlr-randr --output HDMI-A-1 --pos 0,0 --mode 1920x1080
wlr-randr --output DP-1 --pos 1920,0 --mode 1920x1080 --transform 90
waybar &
qwertywm &
qwertywmctl wait-for-socket
. ~/.config/qwertywm/config
qwertywmctl focus-output HDMI-A-1 && qwertywmctl view 1
qwertywmctl focus-output DP-1 && qwertywmctl view 11
EOF
chmod +x ~/.config/river/init

# Iniciar river desde un TTY o DM
river
```

### Ubuntu 24.04

```sh
# Dependencias
sudo apt install golang-go river wayland-protocols libxkbcommon-dev

# Compilar
git clone https://github.com/ldgnu/qwertywm.git
cd qwertywm
go build ./cmd/qwertywm
go build ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/

# (seguir los pasos de config de arriba)
```

### Build rápido (cualquier distro)

```sh
curl -sSL https://github.com/ldgnu/qwertywm/archive/main.tar.gz | tar xz
cd qwertywm-main
go build ./cmd/qwertywm && go build ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/
```

## Configuración

qwertywm se configura con commands de `qwertywmctl` en tu init script.
Ejemplo completo en [`example/init`](example/init) o en la configuración
personal de ldgnu en [`dotfiles`](https://github.com/ldgnu/dotfiles).

## Uso básico

```sh
qwertywmctl focus next        # mover foco a la siguiente ventana
qwertywmctl view 3            # ir al workspace 3
qwertywmctl send 5            # mandar ventana al workspace 5
qwertywmctl cycle-layout monocle,left,top  # cambiar layout
qwertywmctl toggle-float      # flotar/desflotar ventana
qwertywmctl close             # cerrar ventana
qwertywmctl spawn kitty       # abrir terminal
qwertywmctl help              # ver todos los comandos
```

## Estructura del proyecto

| Path | Qué es |
| --- | --- |
| `core/` | State machine: modelo, layouts, comandos. Go puro, sin Wayland. |
| `bridge/` | Adaptador del protocolo river. Conecta el core con el compositor. |
| `ipc/` | Socket unix + JSON: commands, queries, subscriptions. |
| `cmd/qwertywm/` | El binary del WM. |
| `cmd/qwertywmctl/` | El CLI para controlar qwertywm. |
| `wire/` | Cliente Wayland en Go puro, sin cgo. |
| `protocol/` | XMLs del protocolo vendeados. |

## License

MIT. El proyecto original weir es MIT por psanford. Ver [LICENSE](LICENSE).
