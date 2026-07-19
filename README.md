<img width="1914" height="1079" alt="imagen" src="https://github.com/user-attachments/assets/e33b5b7f-71ed-4b55-983e-dc1b6f79dd95" /># qwertywm 🚧

> **⚠️ Estado: desarrollo activo.** Puede tener bugs, cosas rotas y APIs que
> cambian sin aviso. Si te lo bajás, bienvenido, pero no hay garantías de que
> ande en tu setup. PRs y issues bienvenidos.

Un window manager dinámico estilo xmonad para [river](https://codeberg.org/river/river),
escrito en Go. Tiling dinámico, configuración programática vía socket unix,
soporte multi-monitor de primera clase.

<img width="1918" height="1079" alt="imagen" src="https://github.com/user-attachments/assets/44bae52a-df8e-474b-b465-d75aa3746cc5" />
<img width="1913" height="1051" alt="imagen" src="https://github.com/user-attachments/assets/7eeb694a-d812-43a9-8541-29ef05e92f32" />

## Instalación rápida

### AUR (Arch Linux)
```sh
paru -S qwertywm
# o
yay -S qwertywm
```

### Interactive installer (recomendado)

Te pregunta todo: terminal, colores, teclado, monitores... y deja todo listo.

```sh
git clone https://github.com/ldgnu/qwertywm
cd qwertywm
./install.sh
```

### Manual

```sh
sudo pacman -S go river waybar fuzzel kitty foot ttf-liberation wlr-randr
git clone https://github.com/ldgnu/qwertywm
cd qwertywm
go build ./cmd/qwertywm
go build -o qwertywmctl ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/
```

Después copiá las configs de `config/` a `~/.config/`.

---

## Atajos principales

| Tecla | Acción |
|-------|--------|
| `Super+Enter` | Terminal (kitty) |
| `Super+d` | Lanzador (fuzzel) |
| `Super+j/k` | Foco siguiente/anterior |
| `Super+1..0` | Escritorios 1-10 |
| `Alt+1..0` | Escritorios 11-20 |
| `Super+Space` | Cambiar layout |
| `Super+f` | Fullscreen |
| `Super+r` | Recargar config |
| `Super+Escape` | Salir |

---

## Créditos

**qwertywm** es un fork de [weir](https://github.com/psanford/weir) por
[psanford](https://github.com/psanford). El proyecto original fue renombrado
y modificado para uso personal por [ldgnu](https://github.com/ldgnu).

Gracias a psanford por el laburazo del core, la arquitectura limpia y toda la
base sólida.

---

## Requisitos

- [river](https://codeberg.org/river/river) ≥ 0.4 (compositor Wayland)
- Go ≥ 1.21 (solo para compilar)
- `wayland` (protocolos)
- `wayland-protocols`
- `libxkbcommon` (para `keyboard-layout`)

### Dependencias opcionales (para el setup completo)

| Paquete | Uso |
|---------|-----|
| `waybar` | Barra de estado |
| `fuzzel` | Lanzador de apps / menú de temas |
| `wl-clipboard` | Clipboard (wl-copy, wl-paste) |
| `cliphist` | Historial de clipboard |
| `kitty` | Terminal |
| `qutebrowser` | Navegador |
| `pavucontrol` | Control de audio |
| `pamixer` | Volumen desde tecla |
| `playerctl` | Control de reproducción (play/pause, next) |
| `blueman` | Bluetooth (blueman-manager) |
| `bluetuith` | Bluetooth TUI |
| `ncpamixer` | Audio mixer TUI |
| `swaybg` | Wallpaper |
| `hyprlock` | Bloqueo de pantalla |
| `grim` + `slurp` + `swappy` | Capturas de pantalla |
| `copyq` | Clipboard GUI |
| `jq` | Procesar JSON desde scripts |
| `curl` | Clima en waybar |
| `ttf-liberation` | Tipografía retro terminal |
| `fastfetch` | Info del sistema |

---

## Instalación

### Arch Linux / CachyOS

```sh
# Dependencias
sudo pacman -S go river wayland wayland-protocols libxkbcommon

# Compilar e instalar
git clone https://github.com/ldgnu/qwertywm.git
cd qwertywm
./build.sh
sudo cp qwertywm qwertywmctl /usr/local/bin/

# (opcional) Dependencias extras para el setup completo
sudo pacman -S waybar fuzzel wl-clipboard cliphist kitty qutebrowser \
  pavucontrol pamixer playerctl blueman bluetuith ncpamixer swaybg \
  hyprlock grim slurp swappy copyq jq curl ttf-liberation fastfetch
```

### Script rápido (cualquier distro con Go)

```sh
curl -sSL https://github.com/ldgnu/qwertywm/archive/main.tar.gz | tar xz
cd qwertywm-main
./build.sh
sudo cp qwertywm qwertywmctl /usr/local/bin/
```

---

## Configuración

qwertywm se configura con comandos de `qwertywmctl` en tu init script de river.

### Init básico (`~/.config/river/init`)

```sh
#!/bin/sh
export XDG_CURRENT_DESKTOP=qwertywm
export XDG_SESSION_DESKTOP=qwertywm

wlr-randr --output HDMI-A-1 --pos 0,0 --mode 1920x1080
wlr-randr --output DP-1 --pos 1920,0 --mode 1920x1080 --transform 90

swaybg -o HDMI-A-1 -i ~/wallpaper.png -m fill &
swaybg -o DP-1 -i ~/wallpaper.png -m fill &
waybar &
wl-paste --watch cliphist store &
qwertywm &
qwertywmctl wait-for-socket
. ~/.config/qwertywm/config
qwertywmctl focus-output HDMI-A-1 && qwertywmctl view 1
qwertywmctl focus-output DP-1 && qwertywmctl view 11
```

### Config de qwertywm (`~/.config/qwertywm/config`)

Ejemplo completo con binds, apariencia y reglas flotantes:

```sh
#!/bin/sh
mod=Super
terminal=kitty
launcher=fuzzel

# Apariencia
qwertywmctl set border-width 2
qwertywmctl set border-color-focused   0x00aa00
qwertywmctl set border-color-unfocused 0x555555
qwertywmctl set gaps 4 8
qwertywmctl set main-ratio 0.5
qwertywmctl set main-count 1
qwertywmctl set main-location left
qwertywmctl workspace-mode independent

# Lanzadores
qwertywmctl bind $mod+Return  spawn $terminal
qwertywmctl bind $mod+d       spawn $launcher
qwertywmctl bind $mod+q       close
qwertywmctl bind $mod+Escape  exit

# Navegación vim
qwertywmctl bind $mod+h  focus prev
qwertywmctl bind $mod+j  focus next
qwertywmctl bind $mod+k  focus prev
qwertywmctl bind $mod+l  focus next

# Layout
qwertywmctl bind $mod+space  cycle-layout monocle,left,top
qwertywmctl bind $mod+f      toggle-fullscreen

# Mouse
qwertywmctl bind-pointer $mod+Left  move
qwertywmctl bind-pointer $mod+Right resize

# Recargar
qwertywmctl bind $mod+r spawn 'killall qwertywm 2>/dev/null; qwertywm &; qwertywmctl wait-for-socket; . ~/.config/qwertywm/config &'

# Escritorios 1-10
for i in 1 2 3 4 5 6 7 8 9; do
    qwertywmctl bind $mod+$i view $i
    qwertywmctl bind $mod+Shift+$i send $i
done
qwertywmctl bind $mod+0  view 10
qwertywmctl bind $mod+Shift+0  send 10

# Escritorios 11-20 (Alt)
for i in 1 2 3 4 5 6 7 8 9; do
    n=$((i + 10))
    qwertywmctl bind Alt+$i view $n
    qwertywmctl bind Alt+Shift+$i send $n
done
qwertywmctl bind Alt+0  view 20
qwertywmctl bind Alt+Shift+0  send 20
```

---

## Uso de qwertywmctl

```sh
qwertywmctl help                    # Todos los comandos
qwertywmctl focus next              # Siguiente ventana
qwertywmctl focus prev              # Ventana anterior
qwertywmctl view 3                  # Ir a workspace 3
qwertywmctl send 5                  # Ventana al workspace 5
qwertywmctl get state               # Estado JSON
qwertywmctl subscribe               # Eventos en tiempo real
qwertywmctl cycle-layout monocle,left,top  # Layouts
qwertywmctl toggle-float            # Flotar/desflotar
qwertywmctl toggle-fullscreen       # Fullscreen
qwertywmctl close                   # Cerrar ventana
qwertywmctl spawn kitty             # Abrir app
qwertywmctl set gaps 4 8            # Configurar gaps
qwertywmctl workspace-mode independent  # Modo workspaces independientes
```

---

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
| `scripts/` | Scripts de instalación para Arch y Ubuntu. |
| `install.sh` | Instalador interactivo (te pregunta todo). |
| `config/` | Configs de ejemplo (river, waybar, fuzzel, foot). |
| `PKGBUILD` | Paquete AUR. |

---

## Licencia

MIT. Ver [LICENSE](LICENSE). Basado en [weir](https://github.com/psanford/weir) por psanford.
