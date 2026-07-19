#!/bin/sh
ip=$(tailscale status 2>/dev/null | head -1 | awk '{print $1}')
[ -z "$ip" ] && status="↓" || status="${ip}"
echo "{\"text\":\"TS ${status}\"}"
