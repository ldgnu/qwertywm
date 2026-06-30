# qwertywm 🚧

> **⚠️ Estado: desarrollo activo.** Puede tener bugs, cosas rotas y APIs que
> cambian sin aviso. Si te lo bajás, bienvenido, pero no hay garantías de que
> ande en tu setup. PRs y issues bienvenidos.

Un window manager dinámico estilo xmonad para [river](https://codeberg.org/river/river),
escrito en Go. Tiling dinámico, configuración programática vía socket unix,
soporte multi-monitor de primera clase.

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
| `Terminus` (font) | Tipografía retro terminal |
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
  hyprlock grim slurp swappy copyq jq curl terminus-font fastfetch
```

### Ubuntu 24.04

```sh
# Dependencias
sudo apt install golang-go river wayland-protocols libxkbcommon-dev

# Compilar
git clone https://github.com/ldgnu/qwertywm.git
cd qwertywm
./build.sh
sudo cp qwertywm qwertywmctl /usr/local/bin/
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
qwertywmctl keyboard-layout latam

# Lanzadores
qwertywmctl bind $mod+Return  spawn $terminal
qwertywmctl bind $mod+d       spawn $launcher
qwertywmctl bind $mod+e       spawn "kitty --title yazi yazi"
qwertywmctl bind $mod+w       spawn qutebrowser
qwertywmctl bind $mod+q       close
qwertywmctl bind $mod+Escape  exit

# Navegación vim
qwertywmctl bind $mod+h  focus prev
qwertywmctl bind $mod+j  focus next
qwertywmctl bind $mod+k  focus prev
qwertywmctl bind $mod+l  focus next
qwertywmctl bind $mod+f  toggle-fullscreen

# Ventanas
qwertywmctl bind $mod+Shift+space toggle-float
qwertywmctl bind $mod+Shift+Tab   send-to-output next
qwertywmctl bind $mod+Tab         focus-output next

# Workspaces fijos por monitor
# HDMI-A-1: Super+1..0 = 1-10
for i in 1 2 3 4 5 6 7 8 9; do
    qwertywmctl bind $mod+$i spawn "qwertywmctl focus-output HDMI-A-1 && qwertywmctl view $i"
    qwertywmctl bind $mod+Shift+$i spawn "qwertywmctl focus-output HDMI-A-1 && qwertywmctl send $i"
done
qwertywmctl bind $mod+0  spawn "qwertywmctl focus-output HDMI-A-1 && qwertywmctl view 10"
qwertywmctl bind $mod+Shift+0  spawn "qwertywmctl focus-output HDMI-A-1 && qwertywmctl send 10"

# DP-1: Alt+1..0 = 11-20
for i in 1 2 3 4 5 6 7 8 9; do
    n=$((i + 10))
    qwertywmctl bind Alt+$i spawn "qwertywmctl focus-output DP-1 && qwertywmctl view $n"
    qwertywmctl bind Alt+Shift+$i spawn "qwertywmctl focus-output DP-1 && qwertywmctl send $n"
done
qwertywmctl bind Alt+0  spawn "qwertywmctl focus-output DP-1 && qwertywmctl view 20"
qwertywmctl bind Alt+Shift+0  spawn "qwertywmctl focus-output DP-1 && qwertywmctl send 20"

# Multimedia
qwertywmctl bind $mod+z  spawn "playerctl play-pause"
qwertywmctl bind $mod+x  spawn "playerctl next"
qwertywmctl bind XF86AudioRaiseVolume spawn "pamixer -i 5"
qwertywmctl bind XF86AudioLowerVolume spawn "pamixer -d 5"
qwertywmctl bind XF86AudioMute        spawn "pamixer -t"

# Layout
qwertywmctl bind $mod+space  cycle-layout monocle,left,top
qwertywmctl bind $mod+v      cycle-layout left,top

# Resize
qwertywmctl bind $mod+Ctrl+h spawn "qwertywmctl resize horizontal -10"
qwertywmctl bind $mod+Ctrl+l spawn "qwertywmctl resize horizontal 10"
qwertywmctl bind $mod+Ctrl+k spawn "qwertywmctl resize vertical -10"
qwertywmctl bind $mod+Ctrl+j spawn "qwertywmctl resize vertical 10"

# Audio devices
qwertywmctl bind Alt+o       spawn "pactl set-default-sink <sink-name> && notify-send 'Auricular por defecto'"
qwertywmctl bind Alt+Shift+o spawn "pactl set-default-sink <sink-name> && notify-send 'Onboard por defecto'"

# Reglas flotantes
qwertywmctl rule add -app-id pavucontrol float
qwertywmctl rule add -title ncpamixer float
qwertywmctl rule add -title bluetuith float

# Capturas
qwertywmctl bind Print          spawn "grim -g \"$(slurp)\" - | wl-copy"
qwertywmctl bind $mod+Shift+p   spawn "grim -g \"$(slurp)\" - | swappy -f -"

# Mouse
qwertywmctl bind-pointer $mod+Left  move
qwertywmctl bind-pointer $mod+Right resize
```

---

## Themes (365+ colores de Gogh)

El repo incluye un sistema de temas basado en [Gogh](https://gogh-co.github.io/Gogh/)
con 365+ esquemas de color para terminal.

### Dependencias para themes

```sh
sudo pacman -S jq fuzzel zip # Arch
sudo apt install jq fuzzel zip # Ubuntu
```

### Uso

```sh
# Generar todos los themes desde Gogh
~/.config/themes/import-all.sh

# Aplicar un theme
~/.config/themes/apply-theme.sh Dracula
~/.config/themes/apply-theme.sh MonoGreen
~/.config/themes/apply-theme.sh C64

# Menú interactivo con fuzzel
~/.config/themes/theme-switcher.sh

# Siguiente / anterior
~/.config/themes/cycle-theme.sh next
~/.config/themes/cycle-theme.sh prev

# Random
~/.config/themes/cycle-theme.sh random
```

### Atajos de teclado (si usás la config de ldgnu)

| Tecla | Acción |
|-------|--------|
| `Super+t` | Menú de temas (fuzzel) |
| `Super+Shift+t` | Theme random |
| `Super+Shift+,` | Theme anterior |
| `Super+Shift+.` | Theme siguiente |

### Qué cambia al aplicar un theme

- **Kitty** → colores ANSI
- **Waybar** → colores de la barra
- **Fuzzel** → colores del lanzador
- **qutebrowser** → colores del navegador
- **Zen Browser** → userChrome.css (reiniciar Zen)
- **GTK apps** (Thunar, etc) → fondo y acentos
- **qwertywm** → bordes de ventanas
- **Telegram** → se genera .tdesktop-theme (importar manual)

---

## Atajos de teclado (resumen)

### Workspaces

| Tecla | Acción |
|-------|--------|
| `Super+1..0` | Workspace 1-10 en HDMI |
| `Alt+1..0` | Workspace 11-20 en DP |
| `Super+Shift+1..0` | Mover ventana a workspace 1-10 |
| `Alt+Shift+1..0` | Mover ventana a workspace 11-20 |

### Navegación

| Tecla | Acción |
|-------|--------|
| `Super+h/j/k/l` | Focus vim (izq/abajo/arriba/derecha) |
| `Super+Tab` | Cambiar entre monitores |
| `Super+Shift+Tab` | Mover ventana al otro monitor |

### Ventanas

| Tecla | Acción |
|-------|--------|
| `Super+Return` | Abrir terminal (kitty) |
| `Super+q` | Cerrar ventana |
| `Super+f` | Fullscreen |
| `Super+Space` | Ciclar layouts (monocle, left, top) |
| `Super+v` | Ciclar layouts left/top |
| `Super+Shift+Space` | Toggle flotante |
| `Super+Ctrl+h/j/k/l` | Redimensionar ventana |

### Lanzadores

| Tecla | App |
|-------|-----|
| `Super+d` | fuzzel (lanzador) |
| `Super+e` | yazi (file manager) |
| `Super+w` | qutebrowser |
| `Super+c` | Clipboard (cliphist) |
| `Super+u` | ncpamixer (audio TUI) |
| `Super+b` | bluetuith (bluetooth TUI) |

### Multimedia

| Tecla | Acción |
|-------|--------|
| `Super+z` | Play/Pause |
| `Super+x` | Next track |
| Volumen físico | pamixer +/-/mute |

### Sistema

| Tecla | Acción |
|-------|--------|
| `Super+Shift+l` | Bloquear pantalla (hyprlock) |
| `Super+Shift+Escape` | Menú de power (fuzzel) |
| `Super+r` | Recargar qwertywm |
| `Super+Escape` | Salir de river |
| `Print` | Capturar región |
| `Super+Shift+p` | Capturar región → swappy |

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
qwertywmctl keyboard-layout latam   # Layout de teclado
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
| `debian/` | Packaging para Ubuntu/Debian. |
| `build.sh` | Build rápido de los binarios. |

---

## Themes (365+ colores estilo Gogh)

El directorio [`themes/`](themes/) contiene el sistema de temas que usamos
acá. Está basado en los esquemas de [Gogh](https://gogh-co.github.io/Gogh/)
y permite cambiar colores en vivo para kitty, waybar, fuzzel, qutebrowser,
GTK, Zen Browser, Telegram y qwertywm.

```sh
# Importar los 365+ temas (requiere jq y el JSON de Gogh)
./themes/import-all.sh

# Aplicar un tema
./themes/apply-theme.sh Dracula
./themes/apply-theme.sh C64
./themes/apply-theme.sh MonoGreen

# Menú interactivo
./themes/theme-switcher.sh

# Siguiente / anterior
./themes/cycle-theme.sh next
./themes/cycle-theme.sh prev
```

### Dependencias para themes

```sh
# Arch
sudo pacman -S jq fuzzel zip terminus-font

# Ubuntu
sudo apt install jq fuzzel zip fonts-terminus
```

## Licencia

MIT. Ver [LICENSE](LICENSE). Basado en [weir](https://github.com/psanford/weir) por psanford.
