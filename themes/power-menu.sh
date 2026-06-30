#!/bin/sh
choice=$(printf "Shutdown\nReboot\nLock\nLogout\nSuspend" | fuzzel --dmenu --prompt="power: " --width=20 --lines=5)
case "$choice" in
  Shutdown) systemctl poweroff ;;
  Reboot)   systemctl reboot ;;
  Lock)     hyprlock ;;
  Logout)   riverctl exit ;;
  Suspend)  systemctl suspend ;;
esac
