#!/bin/sh
usage=$(cat /sys/class/drm/card*/device/gpu_busy_percent 2>/dev/null)
temp=$(sensors -u amdgpu-pci-0300 2>/dev/null | awk '/edge/{getline; printf "%d", $2; exit}')
echo "{\"text\":\"GPU ${usage}% ${temp}°C\"}"
