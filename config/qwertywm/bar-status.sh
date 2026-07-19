#!/bin/sh
exec 2>/dev/null
qwertywmctl get state | jq -r '
  . as $s |
  ([ $s.outputs[] | select(.focused) | .workspace ][0] // "") as $current |
  ([ $s.outputs[] | select(.focused) | .name ][0] // "") as $out |
  ( if $out == "HDMI-A-1" then [range(1;10)] | map(tostring)
    else [range(10;20)] | map(tostring) end ) as $all |
  [ $s.workspaces[] | select(.windows != null and (.windows | length > 0)) | .name ] as $occupied |
  [ $all[] | . as $ws |
    if $ws == $current then " <span color=\"#000000\" background="'"'"''"$GREEN"'"'"'"> \($ws) </span> "
    elif ([$occupied[] | select(. == $ws)] | length > 0) then " \($ws) "
    else ""
    end
  ] | map(select(. != "")) | join("") as $tags |
  ( [ $s.windows[] | select(.focused) | .title ][0] // "" ) as $title |
  if $title == "" then $tags elif $tags == "" then $title else "\($tags)│ \($title)" end
' 2>/dev/null || echo ""
