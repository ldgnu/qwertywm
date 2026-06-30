config.load_autoconfig()
c.colors.webpage.bg = "$BG"

c.colors.completion.category.bg = "$BG"
c.colors.completion.category.fg = "$GREEN"
c.colors.completion.category.border.bottom = "$GREEN"
c.colors.completion.category.border.top = "$GREEN"
c.colors.completion.fg = "$FG"
c.colors.completion.item.selected.bg = "$GREEN"
c.colors.completion.item.selected.fg = "#000000"
c.colors.completion.item.selected.border.bottom = "$GREEN"
c.colors.completion.item.selected.border.top = "$GREEN"
c.colors.completion.match.fg = "$YELLOW"

c.colors.downloads.bar.bg = "$BG"
c.colors.downloads.start.fg = "#000000"
c.colors.downloads.start.bg = "$GREEN"
c.colors.downloads.stop.fg = "#000000"
c.colors.downloads.stop.bg = "$RED"
c.colors.downloads.error.fg = "#ffffff"
c.colors.downloads.error.bg = "$RED"

c.colors.hints.fg = "#000000"
c.colors.hints.bg = "$YELLOW"

c.colors.keyhint.fg = "$FG"
c.colors.keyhint.bg = "$BG"
c.colors.keyhint.suffix.fg = "$YELLOW"

c.colors.messages.error.fg = "#ffffff"
c.colors.messages.error.bg = "$RED"
c.colors.messages.info.fg = "$FG"
c.colors.messages.info.bg = "$BG"
c.colors.messages.warning.fg = "#000000"
c.colors.messages.warning.bg = "$YELLOW"

c.colors.prompts.fg = "$FG"
c.colors.prompts.bg = "$BG"
c.colors.prompts.selected.fg = "#000000"
c.colors.prompts.selected.bg = "$GREEN"

c.colors.statusbar.normal.fg = "$FG"
c.colors.statusbar.normal.bg = "$BG"
c.colors.statusbar.insert.fg = "#000000"
c.colors.statusbar.insert.bg = "$GREEN"
c.colors.statusbar.command.fg = "$FG"
c.colors.statusbar.command.bg = "$BG"
c.colors.statusbar.command.private.fg = "$FG"
c.colors.statusbar.command.private.bg = "$BG"
c.colors.statusbar.private.fg = "$FG"
c.colors.statusbar.private.bg = "$BG"
c.colors.statusbar.url.fg = "$BLUE"
c.colors.statusbar.url.hover.fg = "$CYAN"
c.colors.statusbar.url.error.fg = "$RED"
c.colors.statusbar.url.success.http.fg = "$FG"
c.colors.statusbar.url.success.https.fg = "$GREEN"
c.colors.statusbar.url.warn.fg = "$YELLOW"

c.colors.tabs.bar.bg = "$BG"
c.colors.tabs.even.bg = "$BG"
c.colors.tabs.even.fg = "$FG"
c.colors.tabs.odd.bg = "$BG_LIGHT"
c.colors.tabs.odd.fg = "$FG"
c.colors.tabs.selected.even.bg = "$GREEN"
c.colors.tabs.selected.even.fg = "#000000"
c.colors.tabs.selected.odd.bg = "$GREEN"
c.colors.tabs.selected.odd.fg = "#000000"

c.fonts.default_family = "Terminus"
c.fonts.default_size = "12pt"

for i in range(1, 9):
    config.unbind(f'<Alt+{i}>')
    config.bind(f'<Ctrl+{i}>', f'tab-focus {i}')
config.unbind('<Alt+9>')
config.bind('<Ctrl+9>', 'tab-focus -1')
config.bind('<Ctrl-Shift-L>', 'spawn --userscript qute-bw')
config.bind('<Ctrl-Shift-X>', 'spawn --userscript qute-bw-lock')
