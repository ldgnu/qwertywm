#!/bin/sh
# Visual test: run weir in a headless river, open a few windows, and
# capture a screenshot of each output. The screenshots land in the
# directory given as $1 (default: ./screenshots).
#
# Set $RIVER, $FOOT, and $GRIM to the binaries to use.
set -eu

outdir="${1:-screenshots}"
RIVER="${RIVER:-river}"
FOOT="${FOOT:-foot}"
GRIM="${GRIM:-grim}"

mkdir -p "$outdir"
outdir="$(cd "$outdir" && pwd)"
dir="$(mktemp -d /tmp/weir-shot.XXXXXX)"
trap 'rm -rf "$dir"' EXIT
mkdir -p -m 0700 "$dir/run"

repo="$(cd "$(dirname "$0")/.." && pwd)"
go build -o "$dir/weir" "$repo/cmd/weir"

cat > "$dir/init" <<EOF
#!/bin/sh
"$dir/weir" -log-level debug 2>"$dir/weir.log" &
sleep 1
"$FOOT" 2>/dev/null &
"$FOOT" 2>/dev/null &
"$FOOT" 2>/dev/null &
sleep 2
"$GRIM" "$outdir/weir-tiled.png" 2>"$dir/grim.log" || true
EOF
chmod +x "$dir/init"

env -i \
    HOME="$dir" \
    PATH="$PATH" \
    XDG_RUNTIME_DIR="$dir/run" \
    WLR_BACKENDS=headless \
    WLR_RENDERER=pixman \
    WLR_LIBINPUT_NO_DEVICES=1 \
    timeout --signal=TERM 6 \
    "$RIVER" -no-xwayland -log-level info -c "$dir/init" >"$dir/river.log" 2>&1 || true
pkill -TERM -f "$dir/weir" 2>/dev/null || true

cat "$dir/grim.log" 2>/dev/null || true
if [ -s "$outdir/weir-tiled.png" ]; then
    echo "wrote $outdir/weir-tiled.png"
else
    echo "FAIL: no screenshot produced"
    cat "$dir/weir.log" 2>/dev/null || true
    exit 1
fi
