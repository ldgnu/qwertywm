#!/bin/sh
# Integration test: run weir inside a real headless river, drive it with
# weirctl, and assert on the JSON state it reports.
#
# Requires river >= 0.4, foot, and jq. Set $RIVER and $FOOT to override
# what is found on PATH (see scripts/fetch-river.sh).
set -eu

RIVER="${RIVER:-river}"
FOOT="${FOOT:-foot}"

dir="$(mktemp -d /tmp/weir-itest.XXXXXX)"
trap 'rm -rf "$dir"' EXIT
mkdir -p -m 0700 "$dir/run"

repo="$(cd "$(dirname "$0")/.." && pwd)"
go build -o "$dir/weir" "$repo/cmd/weir"
go build -o "$dir/weirctl" "$repo/cmd/weirctl"

# The test body runs inside the river session (so WAYLAND_DISPLAY and the
# control socket are reachable). It writes PASS/FAIL lines to the verdict
# file and ends the session with "weirctl exit".
cat > "$dir/test" <<'TESTEOF'
#!/bin/sh
# Runs inside the river session. $WEIR_TEST_DIR and $FOOT come from the
# outer script's environment; everything else is derived here.
set -u
verdict="$WEIR_TEST_DIR/verdict"
ctl="$WEIR_TEST_DIR/weirctl"
echo "test script started (FOOT=$FOOT)" >>"$verdict"

ok() { echo "ok: $1" >>"$verdict"; }
fail() { echo "FAIL: $1" >>"$verdict"; }

# expect <description> <jq-expression-that-must-be-true>
expect() {
    desc="$1"; expr="$2"
    state="$("$ctl" get state 2>>"$verdict")"
    if [ -z "$state" ]; then
        fail "$desc: weirctl get state returned nothing"
        return
    fi
    if printf '%s' "$state" | jq -e "$expr" >/dev/null 2>&1; then
        ok "$desc"
    else
        fail "$desc: jq '$expr' is false. state: $(printf '%s' "$state" | jq -c .)"
    fi
}

"$WEIR_TEST_DIR/weir" -log-level debug 2>"$WEIR_TEST_DIR/weir.log" &
sleep 1

expect "starts with one output and no windows" \
    '.outputs | length == 1'
expect "default workspace 1 is focused" \
    '.outputs[0].workspace == "1"'

"$FOOT" 2>/dev/null &
"$FOOT" 2>/dev/null &
"$FOOT" 2>/dev/null &
sleep 1

expect "three windows appear" '.windows | length == 3'
expect "all windows are visible" '[.windows[].visible] | all'
expect "the newest window is focused" '.windows[-1].focused == true'
expect "windows do not overlap (master is 60% of 1280 minus the 2px border inset)" \
    '.windows[0].width == 764'

"$ctl" set main-ratio 0.25 || fail "set main-ratio"
expect "main-ratio change resizes the master" '.windows[0].width == 316'

"$ctl" set-layout monocle || fail "set-layout"
expect "monocle gives every window the full output" \
    '[.windows[] | select(.width == 1276 and .height == 716)] | length == 3'
"$ctl" set-layout tile || fail "set-layout tile"

"$ctl" focus main || fail "focus main"
expect "focus main focuses the first window" '.windows[0].focused == true'

"$ctl" send 5 || fail "send"
expect "sent window is on workspace 5 and hidden" \
    '(.windows[0].workspace == "5") and (.windows[0].visible == false)'
expect "two windows remain visible" \
    '[.windows[] | select(.visible)] | length == 2'

"$ctl" view 5 || fail "view"
expect "viewing workspace 5 shows the sent window" \
    '(.outputs[0].workspace == "5") and (.windows[0].visible == true)'
expect "windows on workspace 1 are now hidden" \
    '[.windows[] | select(.visible)] | length == 1'

"$ctl" close || fail "close"
sleep 0.5
expect "close removes the window" '.windows | length == 2'

# Key bindings: register, list, and (if wtype is available) actually press.
"$ctl" view 1 || fail "view 1"
"$ctl" bind Super+j focus next || fail "bind"
"$ctl" bind Super+8 view 8 || fail "bind view"
"$ctl" bind-pointer Super+Left move || fail "bind-pointer"
expect "bindings appear in the state" \
    '(.bindings | length == 3) and ([.bindings[].chord] | contains(["Super+j"]))'
if [ -n "${WTYPE:-}" ]; then
    # Inject a real Super+8 key press through a virtual keyboard. It must
    # travel compositor -> xkb binding -> weir -> view 8.
    "$WTYPE" -M logo -P 8 -p 8 -m logo 2>>"$verdict" || fail "wtype"
    sleep 0.5
    expect "pressing Super+8 switches to workspace 8" \
        '.outputs[0].workspace == "8"'
fi
"$ctl" unbind Super+j || fail "unbind"
expect "unbind removes the binding" '.bindings | length == 2'

# Spawn: the child must run inside the session.
"$ctl" spawn "touch $WEIR_TEST_DIR/spawned" || fail "spawn"
sleep 0.5
if [ -f "$WEIR_TEST_DIR/spawned" ]; then
    ok "spawn runs a command"
else
    fail "spawn did not run the command"
fi

echo done >>"$verdict"
"$ctl" exit
TESTEOF
chmod +x "$dir/test"

env -i \
    HOME="$dir" \
    PATH="$PATH" \
    WEIR_TEST_DIR="$dir" \
    FOOT="$FOOT" \
    WTYPE="${WTYPE:-}" \
    XDG_RUNTIME_DIR="$dir/run" \
    WLR_BACKENDS=headless \
    WLR_RENDERER=pixman \
    WLR_LIBINPUT_NO_DEVICES=1 \
    timeout --signal=KILL 30 \
    "$RIVER" -no-xwayland -log-level info -c "$dir/test" >"$dir/river.log" 2>&1 || true
pkill -TERM -f "$dir/weir" 2>/dev/null || true

echo "=== verdict ==="
cat "$dir/verdict" 2>/dev/null || { echo "FAIL: test produced no verdict"; cat "$dir/river.log" | tail -20; exit 1; }
echo
if ! grep -q "^done$" "$dir/verdict"; then
    echo "FAIL: test did not run to completion"
    echo "=== weir log ==="; tail -20 "$dir/weir.log" 2>/dev/null
    exit 1
fi
if grep -q "^FAIL" "$dir/verdict"; then
    echo "=== weir log ==="; tail -20 "$dir/weir.log" 2>/dev/null
    echo "RESULT: FAIL"
    exit 1
fi
echo "RESULT: PASS ($(grep -c '^ok' "$dir/verdict") assertions)"
