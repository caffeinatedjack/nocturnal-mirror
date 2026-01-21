package tui

import (
	bubbletea "github.com/charmbracelet/bubbletea"
)

// Key definitions for TUI navigation and actions.
// Supports both vim-style and arrow key bindings.

type KeyMap struct {
	// Navigation
	Up       string
	Down     string
	Left     string
	Right    string
	Home     string
	End      string
	PageUp   string
	PageDown string

	// Actions
	Enter   string
	Escape  string
	Space   string
	Back    string
	Help    string
	Quit    string
	Refresh string

	// Item actions
	Edit     string
	Create   string
	Delete   string
	Activate string
	Complete string
	Validate string
	Actioned string
	View     string
	Search   string
}

// DefaultKeyMap returns default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up:       "k,up",
		Down:     "j,down",
		Left:     "h,left",
		Right:    "l,right",
		Home:     "home",
		End:      "end",
		PageUp:   "pgup",
		PageDown: "pgdown",

		// Actions
		Enter:   "enter",
		Escape:  "escape,esc,q",
		Space:   " ",
		Back:    "escape,esc,backspace",
		Help:    "?",
		Quit:    "q",
		Refresh: "r",

		// Item actions
		Edit:     "e",
		Create:   "n",
		Delete:   "d",
		Activate: "a",
		Complete: "c",
		Validate: "v",
		Actioned: "x",
		View:     "enter",
		Search:   "/",
	}
}

// Matches checks if a key press matches a key map entry.
func (km KeyMap) Matches(keyMsg bubbletea.KeyMsg, keyStr string) bool {
	if keyStr == "" {
		return false
	}

	alternatives := parseKeyAlternatives(keyStr)
	key := keyMsg.String()

	for _, alt := range alternatives {
		if key == alt {
			return true
		}
	}

	return false
}

// parseKeyAlternatives splits a key string by commas.
func parseKeyAlternatives(keyStr string) []string {
	var alts []string
	current := ""
	for _, r := range keyStr {
		if r == ',' {
			if current != "" {
				alts = append(alts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		alts = append(alts, current)
	}
	return alts
}

// IsQuitKey checks if a key press is a quit action.
func (km KeyMap) IsQuitKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Quit) || km.Matches(keyMsg, km.Escape)
}

// IsHelpKey checks if a key press is a help action.
func (km KeyMap) IsHelpKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Help)
}

// IsRefreshKey checks if a key press is a refresh action.
func (km KeyMap) IsRefreshKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Refresh)
}

// IsUpKey checks if a key press navigates up.
func (km KeyMap) IsUpKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Up)
}

// IsDownKey checks if a key press navigates down.
func (km KeyMap) IsDownKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Down)
}

// IsLeftKey checks if a key press navigates left.
func (km KeyMap) IsLeftKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Left)
}

// IsRightKey checks if a key press navigates right.
func (km KeyMap) IsRightKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Right)
}

// IsEnterKey checks if a key press is enter.
func (km KeyMap) IsEnterKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Enter)
}

// IsEscapeKey checks if a key press is escape.
func (km KeyMap) IsEscapeKey(keyMsg bubbletea.KeyMsg) bool {
	return km.Matches(keyMsg, km.Escape)
}
