#!/bin/sh
if [ -n "$SSH_CLIENT" ]; then
  ip=$(echo "$SSH_CLIENT" | awk '{print $1}')
  echo "{\"text\":\"SSH ${ip}\"}"
elif [ -n "$TMUX" ]; then
  echo "{\"text\":\"TMUX\"}"
else
  echo "{\"text\":\"\"}"
fi
