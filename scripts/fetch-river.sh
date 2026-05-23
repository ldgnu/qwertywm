#!/bin/sh
# Fetch river and foot from the nixpkgs binary cache for integration
# testing, without touching the host system. Prints the env vars to pass to
# scripts/smoke-test.sh.
#
# Usage: eval "$(scripts/fetch-river.sh)"
set -eu

out="$(nix-build --no-out-link -I nixpkgs=channel:nixos-unstable \
    -E 'with import <nixpkgs> {}; [ river foot ]' 2>/dev/null)"
river_path="$(echo "$out" | grep -- '-river-')"
foot_path="$(echo "$out" | grep -- '-foot-')"
echo "export RIVER=$river_path/bin/river"
echo "export FOOT=$foot_path/bin/foot"
