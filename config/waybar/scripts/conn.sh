#!/bin/sh
ssh_count=$(ps aux | grep -cE "[s]sh .*@")
tmux_count=$(ps aux | grep -c "[t]mux")

if [ "$ssh_count" -gt 0 ]; then
  echo "{\"text\":\"SSH $ssh_count\"}"
elif [ "$tmux_count" -gt 0 ]; then
  echo "{\"text\":\"TMUX\"}"
else
  echo "{\"text\":\"\"}"
fi
