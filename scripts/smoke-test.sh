#!/bin/sh
# Smoke test: run weir inside a real headless river compositor and verify
# that windows get managed.
#
# Requires river >= 0.4 and foot on PATH (or set $RIVER and $FOOT, e.g. to
# store paths produced by scripts/fetch-river.sh). No GPU, no display, and
# no seat are needed: wlroots' headless backend and pixman renderer do
# everything in software.
#
# Usage: scripts/smoke-test.sh [seconds-to-run]
set -eu

duration="${1:-4}"
RIVER="${RIVER:-river}"
FOOT="${FOOT:-foot}"

dir="$(mktemp -d /tmp/weir-smoke.XXXXXX)"
trap 'rm -rf "$dir"' EXIT
mkdir -p -m 0700 "$dir/run"

repo="$(cd "$(dirname "$0")/.." && pwd)"
go build -o "$dir/weir" "$repo/cmd/weir"

# The init script river runs once its Wayland socket is ready. It starts
# weir and some clients in the background and exits; river survives until
# the outer timeout kills it, at which point river SIGTERMs the init
# process group, taking weir and the clients with it.
cat > "$dir/init" <<EOF
#!/bin/sh
"$dir/weir" -log-level debug 2>"$dir/weir.log" &
sleep 1
"$FOOT" 2>>"$dir/foot1.log" &
"$FOOT" 2>>"$dir/foot2.log" &
"$FOOT" 2>>"$dir/foot3.log" &
EOF
chmod +x "$dir/init"

echo "=== running river (headless) for ${duration}s ==="
env -i \
    HOME="$dir" \
    PATH="$PATH" \
    XDG_RUNTIME_DIR="$dir/run" \
    WLR_BACKENDS=headless \
    WLR_RENDERER=pixman \
    WLR_LIBINPUT_NO_DEVICES=1 \
    timeout --signal=TERM "$duration" \
    "$RIVER" -no-xwayland -log-level debug -c "$dir/init" >"$dir/river.log" 2>&1 || true
# Belt and suspenders: nothing from this run may outlive the test.
pkill -TERM -f "$dir/weir" 2>/dev/null || true

echo "=== weir log ==="
cat "$dir/weir.log" 2>/dev/null || echo "(weir produced no log)"
echo
echo "=== river window management log ==="
grep -iE "error|wm\)|posted" "$dir/river.log" | grep -vi "xdg-activation" | tail -30 || true
echo

# Verdict.
fail=0
if ! grep -q "weir started" "$dir/weir.log" 2>/dev/null; then
    echo "FAIL: weir did not start"
    fail=1
fi
windows=$(grep -c "window added" "$dir/weir.log" 2>/dev/null || true)
if [ "${windows:-0}" -lt 3 ]; then
    echo "FAIL: weir saw ${windows:-0} windows, expected 3"
    fail=1
fi
mapped=$(grep -c "mapped" "$dir/river.log" 2>/dev/null || true)
if [ "${mapped:-0}" -lt 3 ]; then
    echo "FAIL: river mapped ${mapped:-0} windows, expected 3 (windows only map once the WM proposes dimensions and finishes a render sequence)"
    fail=1
fi
if grep -qiE "level=ERROR|protocol error" "$dir/weir.log" 2>/dev/null; then
    echo "FAIL: weir reported an error"
    fail=1
fi
if grep -qi "error in client communication\|posted .*error" "$dir/river.log"; then
    echo "FAIL: river reported a client protocol error"
    fail=1
fi
if [ "$fail" = 0 ]; then
    echo "PASS: weir managed $windows windows (all mapped) in a real river session"
fi
exit $fail
