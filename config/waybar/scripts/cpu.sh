#!/bin/sh
usage=$(grep 'cpu ' /proc/stat | awk '{print ($2+$4)*100/($2+$4+$5)}' | cut -d. -f1)
temp=$(sensors -u coretemp-isa-0000 2>/dev/null | awk '/temp1_input/{printf "%d", $2; exit}')
echo "{\"text\":\"CPU ${usage}% ${temp}°C\"}"
