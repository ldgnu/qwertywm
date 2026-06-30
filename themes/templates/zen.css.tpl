:root {
  --tty-bg: $BG;
  --tty-fg: $FG;
  --tty-green: $GREEN;
  --tty-red: $RED;
  --tty-yellow: $YELLOW;
  --tty-blue: $BLUE;
  --tty-cyan: $CYAN;
}
* { scrollbar-color: var(--tty-green) var(--tty-bg) !important; }
#navigator-toolbox, #nav-bar, #TabsToolbar, #PersonalToolbar,
#urlbar, #sidebar-box, #sidebar-header, #sidebar-search-container,
panel, menupopup, findbar, .findbar-container {
  background: var(--tty-bg) !important;
  color: var(--tty-fg) !important;
  border-color: var(--tty-green) !important;
}
#urlbar { border: 1px solid var(--tty-green) !important; }
#urlbar-input { color: var(--tty-green) !important; background: var(--tty-bg) !important; }
.tabbrowser-tab { background: var(--tty-bg) !important; color: var(--tty-fg) !important; }
.tabbrowser-tab[selected] { background: var(--tty-green) !important; color: #000000 !important; }
.tabbrowser-tab:hover { background: $BG_LIGHT !important; }
#statuspanel-label { background: var(--tty-bg) !important; color: var(--tty-green) !important; border: 1px solid var(--tty-green) !important; }
.autocomplete-richlistitem { color: var(--tty-fg) !important; background: var(--tty-bg) !important; }
.autocomplete-richlistitem[selected] { background: var(--tty-green) !important; color: #000000 !important; }
menuitem, menuitem[selected] { color: var(--tty-fg) !important; }
menuitem[selected] { background: var(--tty-green) !important; color: #000000 !important; }
toolbarbutton { color: var(--tty-green) !important; fill: var(--tty-green) !important; }
toolbarbutton:hover { background: $BG_LIGHT !important; }
.urlbarView-row { color: var(--tty-fg) !important; }
.urlbarView-row[selected] { background: var(--tty-green) !important; color: #000000 !important; }
#identity-box { color: var(--tty-green) !important; }
