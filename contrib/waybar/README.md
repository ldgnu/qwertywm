# waybar integration

river 0.4 removed `river-status-unstable-v1`, so waybar's built-in
`river/tags` and `river/window` modules do not work with river 0.4 or weir.
Feed a `custom` module from weir's control socket instead.

Install `weir-workspaces` somewhere on your `PATH` (it needs `weirctl` and
`jq`), then add to `~/.config/waybar/config`:

```json
"modules-left": ["custom/weir"],
"custom/weir": {
    "exec": "weir-workspaces",
    "return-type": "json",
    "escape": false,
    "tooltip": true
}
```

The module renders the focused workspace in bold brackets and lists every
visible or occupied workspace. It updates on every weir state change (the
subscription delivers the current state immediately on connect, and only the
latest state if waybar falls behind) and reconnects automatically if weir
restarts.

The same `weirctl subscribe` stream carries the full state snapshot — the
focused window title, layouts, per-output workspaces — so a window-title
module or anything else is a different `jq` expression away. `weirctl get
state | jq .` shows everything available.
