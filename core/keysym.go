package core

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Keysym is an xkbcommon keysym value.
type Keysym uint32

// Modifiers is the river_seat_v1.modifiers bitfield.
type Modifiers uint32

const (
	ModShift Modifiers = 1
	ModCtrl  Modifiers = 4
	ModAlt   Modifiers = 8 // mod1
	ModMod3  Modifiers = 32
	ModSuper Modifiers = 64 // mod4, the "logo" key
	ModMod5  Modifiers = 128
)

// modifierNames maps the accepted modifier spellings to their bits.
// Matching is case-insensitive.
var modifierNames = map[string]Modifiers{
	"shift":   ModShift,
	"ctrl":    ModCtrl,
	"control": ModCtrl,
	"alt":     ModAlt,
	"mod1":    ModAlt,
	"mod3":    ModMod3,
	"super":   ModSuper,
	"logo":    ModSuper,
	"mod4":    ModSuper,
	"mod5":    ModMod5,
	"none":    0,
}

// String renders the modifiers in canonical order, e.g. "Super+Shift".
func (m Modifiers) String() string {
	if m == 0 {
		return "None"
	}
	var parts []string
	for _, p := range []struct {
		bit  Modifiers
		name string
	}{
		{ModSuper, "Super"}, {ModCtrl, "Ctrl"}, {ModAlt, "Alt"},
		{ModShift, "Shift"}, {ModMod3, "Mod3"}, {ModMod5, "Mod5"},
	} {
		if m&p.bit != 0 {
			parts = append(parts, p.name)
		}
	}
	return strings.Join(parts, "+")
}

// ParseChord parses a key chord like "Super+Shift+Return" or "Super+j" into
// modifiers and a keysym. The final +-separated component is the key; the
// rest are modifiers. "None+x" binds x with no modifiers.
func ParseChord(s string) (Modifiers, Keysym, error) {
	parts := strings.Split(s, "+")
	if len(parts) == 0 || parts[len(parts)-1] == "" {
		return 0, 0, cmdErrf("invalid key chord %q", s)
	}
	var mods Modifiers
	for _, p := range parts[:len(parts)-1] {
		bit, ok := modifierNames[strings.ToLower(p)]
		if !ok {
			return 0, 0, cmdErrf("unknown modifier %q in %q (want Super, Ctrl, Alt, Shift, Mod3, Mod5, or None)", p, s)
		}
		mods |= bit
	}
	sym, err := KeysymFromName(parts[len(parts)-1])
	if err != nil {
		return 0, 0, err
	}
	return mods, sym, nil
}

// KeysymFromName resolves a key name to an xkbcommon keysym.
//
// Resolution order:
//  1. The named-key table (Return, Escape, F1, XF86AudioMute, ...),
//     case-insensitively.
//  2. A single character: printable ASCII and Latin-1 map to themselves
//     (with uppercase ASCII letters lowered to their unshifted keysym);
//     other Unicode characters map to 0x01000000 | codepoint per the
//     xkbcommon convention.
func KeysymFromName(name string) (Keysym, error) {
	if name == "" {
		return 0, cmdErrf("empty key name")
	}
	if sym, ok := namedKeysyms[strings.ToLower(name)]; ok {
		return sym, nil
	}
	if utf8.RuneCountInString(name) == 1 {
		r, _ := utf8.DecodeRuneInString(name)
		// Bindings should use the unshifted keysym: the J key with Super
		// held produces keysym "j", not "J". Accept either spelling.
		if r >= 'A' && r <= 'Z' {
			r = unicode.ToLower(r)
		}
		switch {
		case r >= 0x20 && r <= 0x7e, r >= 0xa0 && r <= 0xff:
			return Keysym(r), nil
		case r >= 0x100:
			return Keysym(0x01000000 | r), nil
		}
	}
	return 0, cmdErrf("unknown key name %q", name)
}

// KeysymName returns the canonical name for a keysym, for display.
func KeysymName(sym Keysym) string {
	if name, ok := keysymToName[sym]; ok {
		return name
	}
	if sym >= 0x20 && sym <= 0x7e || sym >= 0xa0 && sym <= 0xff {
		return string(rune(sym))
	}
	if sym&0x01000000 != 0 {
		return string(rune(sym &^ 0x01000000))
	}
	return fmt.Sprintf("0x%x", uint32(sym))
}

// namedKeysyms maps lowercase key names to keysym values. The values come
// from xkbcommon-keysyms.h. This is the subset that covers real-world
// window manager bindings; single characters are handled separately by
// KeysymFromName.
var namedKeysyms = map[string]Keysym{
	// TTY function keys.
	"backspace": 0xff08,
	"tab":       0xff09,
	"return":    0xff0d,
	"enter":     0xff0d,
	"pause":     0xff13,
	"escape":    0xff1b,
	"delete":    0xffff,
	// Motion.
	"home":      0xff50,
	"left":      0xff51,
	"up":        0xff52,
	"right":     0xff53,
	"down":      0xff54,
	"prior":     0xff55,
	"page_up":   0xff55,
	"next":      0xff56,
	"page_down": 0xff56,
	"end":       0xff57,
	"insert":    0xff63,
	"menu":      0xff67,
	// Keypad.
	"kp_enter":    0xff8d,
	"kp_home":     0xff95,
	"kp_left":     0xff96,
	"kp_up":       0xff97,
	"kp_right":    0xff98,
	"kp_down":     0xff99,
	"kp_prior":    0xff9a,
	"kp_next":     0xff9b,
	"kp_end":      0xff9c,
	"kp_insert":   0xff9e,
	"kp_delete":   0xff9f,
	"kp_multiply": 0xffaa,
	"kp_add":      0xffab,
	"kp_subtract": 0xffad,
	"kp_decimal":  0xffae,
	"kp_divide":   0xffaf,
	"kp_0":        0xffb0,
	"kp_1":        0xffb1,
	"kp_2":        0xffb2,
	"kp_3":        0xffb3,
	"kp_4":        0xffb4,
	"kp_5":        0xffb5,
	"kp_6":        0xffb6,
	"kp_7":        0xffb7,
	"kp_8":        0xffb8,
	"kp_9":        0xffb9,
	// Function keys.
	"f1": 0xffbe, "f2": 0xffbf, "f3": 0xffc0, "f4": 0xffc1,
	"f5": 0xffc2, "f6": 0xffc3, "f7": 0xffc4, "f8": 0xffc5,
	"f9": 0xffc6, "f10": 0xffc7, "f11": 0xffc8, "f12": 0xffc9,
	"f13": 0xffca, "f14": 0xffcb, "f15": 0xffcc, "f16": 0xffcd,
	"f17": 0xffce, "f18": 0xffcf, "f19": 0xffd0, "f20": 0xffd1,
	"f21": 0xffd2, "f22": 0xffd3, "f23": 0xffd4, "f24": 0xffd5,
	// Common named printables (riverctl-style spellings).
	"space":        0x0020,
	"exclam":       0x0021,
	"quotedbl":     0x0022,
	"numbersign":   0x0023,
	"dollar":       0x0024,
	"percent":      0x0025,
	"ampersand":    0x0026,
	"apostrophe":   0x0027,
	"parenleft":    0x0028,
	"parenright":   0x0029,
	"asterisk":     0x002a,
	"plus":         0x002b,
	"comma":        0x002c,
	"minus":        0x002d,
	"period":       0x002e,
	"slash":        0x002f,
	"colon":        0x003a,
	"semicolon":    0x003b,
	"less":         0x003c,
	"equal":        0x003d,
	"greater":      0x003e,
	"question":     0x003f,
	"at":           0x0040,
	"bracketleft":  0x005b,
	"backslash":    0x005c,
	"bracketright": 0x005d,
	"asciicircum":  0x005e,
	"underscore":   0x005f,
	"grave":        0x0060,
	"braceleft":    0x007b,
	"bar":          0x007c,
	"braceright":   0x007d,
	"asciitilde":   0x007e,
	// XF86 media and laptop keys.
	"xf86audiolowervolume":  0x1008ff11,
	"xf86audiomute":         0x1008ff12,
	"xf86audioraisevolume":  0x1008ff13,
	"xf86audioplay":         0x1008ff14,
	"xf86audiostop":         0x1008ff15,
	"xf86audioprev":         0x1008ff16,
	"xf86audionext":         0x1008ff17,
	"xf86audiomicmute":      0x1008ffb2,
	"xf86monbrightnessup":   0x1008ff02,
	"xf86monbrightnessdown": 0x1008ff03,
	"xf86display":           0x1008ff59,
	"xf86search":            0x1008ff1b,
	"xf86explorer":          0x1008ff5d,
	"xf86calculator":        0x1008ff1d,
	"xf86mail":              0x1008ff19,
	"xf86www":               0x1008ff2e,
	"xf86homepage":          0x1008ff18,
	"xf86sleep":             0x1008ff2f,
	"xf86poweroff":          0x1008ff2a,
	"xf86eject":             0x1008ff2c,
	"print":                 0xff61,
	"sys_req":               0xff15,
	"scroll_lock":           0xff14,
	"num_lock":              0xff7f,
	"caps_lock":             0xffe5,
}

// keysymToName is the reverse mapping with canonical capitalization for
// display. Built once at init from a canonical-name list.
var keysymToName = map[Keysym]string{}

func init() {
	// Prefer the riverctl-style canonical spellings for display.
	canonical := []string{
		"BackSpace", "Tab", "Return", "Pause", "Escape", "Delete",
		"Home", "Left", "Up", "Right", "Down", "Page_Up", "Page_Down",
		"End", "Insert", "Menu", "Print", "space", "comma", "period",
		"minus", "equal", "slash", "backslash", "semicolon", "apostrophe",
		"bracketleft", "bracketright", "grave",
		"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10",
		"F11", "F12",
		"XF86AudioLowerVolume", "XF86AudioMute", "XF86AudioRaiseVolume",
		"XF86AudioPlay", "XF86AudioNext", "XF86AudioPrev",
		"XF86MonBrightnessUp", "XF86MonBrightnessDown",
	}
	for _, name := range canonical {
		if sym, ok := namedKeysyms[strings.ToLower(name)]; ok {
			if _, exists := keysymToName[sym]; !exists {
				keysymToName[sym] = name
			}
		}
	}
}

// PointerButtons maps button names to Linux input event codes for pointer
// bindings.
var pointerButtons = map[string]uint32{
	"left":    0x110, // BTN_LEFT
	"right":   0x111, // BTN_RIGHT
	"middle":  0x112, // BTN_MIDDLE
	"side":    0x113, // BTN_SIDE
	"extra":   0x114, // BTN_EXTRA
	"forward": 0x115, // BTN_FORWARD
	"back":    0x116, // BTN_BACK
	"task":    0x117, // BTN_TASK
}

// ParsePointerChord parses a pointer chord like "Super+Left" into modifiers
// and a Linux button code.
func ParsePointerChord(s string) (Modifiers, uint32, error) {
	parts := strings.Split(s, "+")
	if len(parts) == 0 || parts[len(parts)-1] == "" {
		return 0, 0, cmdErrf("invalid pointer chord %q", s)
	}
	var mods Modifiers
	for _, p := range parts[:len(parts)-1] {
		bit, ok := modifierNames[strings.ToLower(p)]
		if !ok {
			return 0, 0, cmdErrf("unknown modifier %q in %q", p, s)
		}
		mods |= bit
	}
	btn, ok := pointerButtons[strings.ToLower(parts[len(parts)-1])]
	if !ok {
		return 0, 0, cmdErrf("unknown pointer button %q (want %s)", parts[len(parts)-1], strings.Join(buttonNames(), ", "))
	}
	return mods, btn, nil
}

// ButtonName returns the name of a Linux button code, for display.
func ButtonName(code uint32) string {
	for name, c := range pointerButtons {
		if c == code {
			return strings.ToUpper(name[:1]) + name[1:]
		}
	}
	return fmt.Sprintf("0x%x", code)
}

func buttonNames() []string {
	names := make([]string, 0, len(pointerButtons))
	for n := range pointerButtons {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
