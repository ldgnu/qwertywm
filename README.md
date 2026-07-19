# qwertywm




Dynamic tiling window manager for the [River](https://codeberg.org/river/river) Wayland compositor.
Forked from [weir](https://github.com/psanford/weir).

<img width="1918" alt="qwertywm escritorios" src="https://github.com/user-attachments/assets/44bae52a-df8e-474b-b465-d75aa3746cc5" />
<img width="1914" alt="qwertywm" src="https://github.com/user-attachments/assets/e33b5b7f-71ed-4b55-983e-dc1b6f79dd95" />
<img width="1913" alt="qwertywm barra" src="https://github.com/user-attachments/assets/7eeb694a-d812-43a9-8541-29ef05e92f32" />


## ⚡ Quick Install (Arch Linux)

### Desde AUR (recomendado)
```bash
paru -S qwertywm
# o con yay:
yay -S qwertywm
```

### Manual
```bash
git clone https://github.com/ldgnu/qwertywm
cd qwertywm
./setup.sh
```

El `setup.sh` instala todo: dependencias, compila, copia configs, y agrega la sesión al display manager.

---

## 📋 Requisitos

- **Arch Linux** (los paquetes están en pacman/AUR)
- **River** 0.4+ (compositor Wayland)
- **kitty** (terminal)
- **fuzzel** (lanzador de aplicaciones)
- **waybar** (barra de estado)
- **foot** (terminal alternativa)
- **wlr-randr** (configuración de monitores)
- **ttf-liberation** (fuente TTY)
- **Go** (para compilar)

---

## 🚀 Cómo usar

### Desde display manager (GDM, SDDM, LightDM, etc.)
1. Cerrar sesión
2. Seleccionar **qwertywm** del menú de sesiones
3. Iniciar sesión

### Desde TTY
```bash
XDG_RUNTIME_DIR=/run/user/$(id -u) river
```

---

## 🎮 Atajos

| Atajo | Acción |
|-------|--------|
| **Super+Enter** | Terminal (kitty) |
| **Super+d** | Lanzador (fuzzel) |
| **Super+q** | Cerrar ventana |
| **Super+Escape** | Salir de River |

### Navegación (vim)
| Atajo | Acción |
|-------|--------|
| **Super+j** | Foco siguiente ventana |
| **Super+k** | Foco anterior ventana |
| **Super+h** | Foco anterior (alternativo) |
| **Super+l** | Foco siguiente (alternativo) |

### Layout
| Atajo | Acción |
|-------|--------|
| **Super+Space** | Ciclar layout (monocle, left, top) |
| **Super+f** | Fullscreen |

### Escritorios
| Atajo | Acción |
|-------|--------|
| **Super+1..0** | Escritorios 1-10 |
| **Super+Shift+1..0** | Mandar ventana a escritorio 1-10 |
| **Alt+1..0** | Escritorios 11-20 |
| **Alt+Shift+1..0** | Mandar ventana a escritorio 11-20 |

### Monitores
| Atajo | Acción |
|-------|--------|
| **Super+w** | Focus DP-1 (izquierda) |
| **Super+e** | Focus HDMI-A-1 (derecha) |
| **Super+Shift+h** | Mandar ventana a DP-1 |
| **Super+Shift+l** | Mandar ventana a HDMI |

### Mouse
| Atajo | Acción |
|-------|--------|
| **Super+Click izquierdo** | Mover ventana |
| **Super+Click derecho** | Redimensionar |

---

## ⚙️ Configuración

Todo se configura en un solo archivo:

```bash
nano ~/.config/qwertywm/config
```

Después de editar, recargar con **Super+r** (mata el proceso y lo reinicia con la nueva config).

### Colores
Estilo TTY clásico: fondo negro, texto gris, bordes verdes.
Todo configurable en `~/.config/qwertywm/config`.

### Barra (waybar)
- Izquierda: números de escritorio
- Centro: título de la ventana
- Derecha: hora AM/PM y fecha

Config en `~/.config/waybar/config`.

---

## 🏗️ Build manual

```bash
git clone https://github.com/ldgnu/qwertywm
cd qwertywm
go build ./cmd/qwertywm
go build -o qwertywmctl ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/
```

Luego copiar configs de `config/` a `~/.config/`.

---

## 📁 Estructura

```
qwertywm/
├── cmd/
│   ├── qwertywm/        # WM principal
│   ├── qwertywmctl/     # CLI de control
│   └── wmsim/           # Simulador ASCII
├── core/                # Motor (modelo, layouts, comandos)
├── bridge/              # Adaptador protocolo River
├── ipc/                 # Socket de control Unix
├── wire/                # Protocolo Wayland puro
├── protocol/            # XML del protocolo
├── config/              # Configuraciones de ejemplo
│   ├── river/           # Init de River
│   ├── qwertywm/        # Atajos y settings
│   ├── waybar/          # Barra de estado
│   ├── fuzzel/          # Lanzador
│   └── foot/            # Terminal alternativa
├── setup.sh             # Instalador automático
└── PKGBUILD             # Paquete AUR
```

---

## 📦 AUR

El PKGBUILD está en el repo. Para mantenerlo actualizado:

```bash
git clone ssh://aur@aur.archlinux.org/qwertywm.git
cd qwertywm
cp ~/Proyectos/qwertywm/PKGBUILD .
cp ~/Proyectos/qwertywm/.SRCINFO .
makepkg --printsrcinfo > .SRCINFO
git add -A && git commit -m "update" && git push
```

---

## 🧪 Tests

```bash
go test ./...
```

---

## 📝 Licencia

MIT
