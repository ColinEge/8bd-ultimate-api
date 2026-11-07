package keydefs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Key describes a bindable action and its associated metadata.
//
// For keyboard actions (page 0x07), a PhysicalKey may exist and can be
// converted to or from the ActionCode using the firmware keycode rules.
// For actions on other pages (mouse, consumer controls, etc.), there is
// typically no corresponding physical key.
type Key struct {
	KeyID  KeyID      // Firmware physical key identifier
	Action ActionCode // HID usage or firmware action

	KeyLabel    string // Printed label on keycap
	ActionLabel string // Human-readable action name

	Page     byte        //
	Category KeyCategory //
}

type KeyID byte

func (k KeyID) String() string {
	return fmt.Sprintf("0x%02x", byte(k))
}

// KeyCategory defines logical groupings of keys for display or mapping purposes.
// Mirrors the categories used in the official software.
type KeyCategory string

const (
	CategoryAlphanumeric KeyCategory = "Alphanumeric"   // Letters, digits
	CategoryFunction     KeyCategory = "Function"       // F1-F12, Fn combos
	CategoryNumpad       KeyCategory = "Numpad"         // Numeric keypad keys
	CategoryNavigation   KeyCategory = "Navigation"     // Arrows, Home, End, PgUp, PgDn
	CategoryModifier     KeyCategory = "Modifier"       // Shift, Ctrl, Alt, Meta
	CategorySymbol       KeyCategory = "Symbol"         // Punctuation keys
	CategoryMouse        KeyCategory = "Mouse Function" // Mouse-like keys (some keyboards map buttons)
	CategoryShortcut     KeyCategory = "Shortcut"       // Special shortcuts (e.g., media, volume)
	CategoryDisable      KeyCategory = "Disable"        // Keys not mapped or ignored

	// Physical keys that exist on the keyboard but cannot be assigned
	// as a bindable usage. These keys generate events but cannot be
	// produced as a mapping target.
	CategoryKeyOnly KeyCategory = "KeyOnly"
)

// KeyEntry represents the YAML-mapped data for a key.
type KeyEntry struct {
	KeyLabel    string `yaml:"keyLabel"`    // Physical key label
	ActionLabel string `yaml:"actionLabel"` // Semantic usage label
}

// CategoryMap holds all key definitions grouped by category.
// Outer map key = category name
// Inner map key = ActionCode (e.g., 0x07 0x00 0x04)
type CategoryMap map[KeyCategory]map[ActionCode]KeyEntry

var (
	// KeyMap is the in-memory store of all key definitions loaded from YAML.
	KeyMap   CategoryMap
	KeyIndex map[ActionCode]IndexedKey

	CategoryOrder []KeyCategory
	ActionOrder   map[KeyCategory][]ActionCode
)

type IndexedKey struct {
	Entry    KeyEntry
	Category KeyCategory
}

// LoadKeyMap reads a YAML file and unmarshals it into KeyMap.
// The YAML should follow the structure: category -> ActionCode -> {hwLabel, usageLabel}.
func LoadKeyMap(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return err
	}

	KeyMap = make(CategoryMap)
	KeyIndex = make(map[ActionCode]IndexedKey)
	ActionOrder = make(map[KeyCategory][]ActionCode)
	CategoryOrder = nil

	top := root.Content[0]

	for i := 0; i < len(top.Content); i += 2 {
		category := KeyCategory(top.Content[i].Value)
		CategoryOrder = append(CategoryOrder, category)

		keyNode := top.Content[i+1]

		KeyMap[category] = make(map[ActionCode]KeyEntry)

		for j := 0; j < len(keyNode.Content); j += 2 {
			acStr := keyNode.Content[j].Value

			ac, err := parseActionCodeString(acStr)
			if err != nil {
				return fmt.Errorf("invalid action code %q: %w", acStr, err)
			}

			var entry KeyEntry
			if err := keyNode.Content[j+1].Decode(&entry); err != nil {
				return err
			}

			KeyMap[category][ac] = entry
			KeyIndex[ac] = IndexedKey{
				Entry:    entry,
				Category: category,
			}
			ActionOrder[category] = append(ActionOrder[category], ac)
		}
	}
	return nil
}

// KeyFromActionCode looks up a Key struct from an ActionCode.
// Returns a fully populated Key with HWKey, labels, page, and category.
// If the usage is not found in the map, a fallback Key is returned with CategoryDisable.
func KeyFromActionCode(action ActionCode) (*Key, bool) {
	idx, exists := KeyIndex[action]
	physicalKey, hasHW := ActionToPhysicalKey(action)

	if !exists {
		return &Key{
			KeyID:       0,
			Action:      action,
			KeyLabel:    "",
			ActionLabel: action.String(),
			Page:        action[0],
			Category:    CategoryDisable,
		}, false
	}

	return &Key{
		KeyID: func() KeyID {
			if hasHW {
				return physicalKey
			}
			return 0
		}(),
		Action:      action,
		KeyLabel:    idx.Entry.KeyLabel,
		ActionLabel: idx.Entry.ActionLabel,
		Page:        action[0],
		Category:    idx.Category,
	}, true
}

// Returns a list of keys that could physically exist on a keybaord
func PhysicalKeys() []*Key {
	var keys []*Key
	for ac, iKey := range KeyIndex {
		// There's no 0x00, it's a modified key, not a keyboard key
		if ac.Len() != minActionCodeLen || (ac[1] != 0x00 && ac[2] != 0x00) {
			continue
		}
		pKey, ok := ActionToPhysicalKey(ac)
		if !ok {
			continue
		}

		keys = append(keys, &Key{
			KeyID:       pKey,
			Action:      ac,
			KeyLabel:    iKey.Entry.KeyLabel,
			ActionLabel: iKey.Entry.ActionLabel,
			Page:        ac[0],
			Category:    iKey.Category,
		})
	}
	return keys
}
