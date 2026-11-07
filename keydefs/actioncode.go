package keydefs

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ActionCode represents an action that can be assigned to a key binding.
//
// Most actions correspond to HID usages, but the keyboard firmware also
// defines several device-specific actions (such as mouse scrolling).
// Because of this, an ActionCode uses a six-byte array with zero padding
// rather than a fixed 3-byte structure.
//
// Keyboard actions (HID page 0x07) follow a structured 3-byte format:
//
//	[page]-[mod]-[key]
//
//	page : HID usage page (always 0x07 for keyboard keys)
//	mod  : modifier usage ID (if present)
//	key  : standard key usage ID
//
// Examples:
//
//	07-00-06  -> C
//	07-e0-00  -> Left Control
//	07-e0-06  -> Left Control + C
//
// Actions from other pages (mouse functions, consumer controls, shortcuts,
// etc.) do not necessarily follow this structure and may contain additional
// bytes defined by the keyboard firmware. These should be treated as opaque
// action identifiers rather than parsed HID usages.
type ActionCode [maxActionCodeLen]byte

const minActionCodeLen = 3
const maxActionCodeLen = 6

// Length returns the actual number of bytes used in the code
func (a ActionCode) Len() int {
	for i := maxActionCodeLen - 1; i >= 0; i-- {
		if a[i] != 0 {
			if i <= minActionCodeLen-1 {
				return minActionCodeLen
			}
			return i + 1
		}
	}
	return 0
}

// String returns the canonical hex representation of an ActionCode.
// Bytes are formatted as lowercase hex separated by dashes.
//
// Examples:
//
//	[0x07 0x00 0x04] -> "07-00-04"
//	[0x01 0x00 0x00 0x00 0xfe] -> "01-00-00-00-fe"
func (a ActionCode) String() string {
	n := a.Len()

	// Keyboard usages must always print 3 bytes
	if n < minActionCodeLen && a[0] == hidPageKeyboard {
		n = minActionCodeLen
	}

	var b []string
	for i := 0; i < n; i++ {
		b = append(b, fmt.Sprintf("%02x", a[i]))
	}

	return strings.Join(b, "-")
}

// ParseActionCode converts a dash-separated hex string into an ActionCode.
//
// Examples:
//
//	"07-00-04"       -> {0x07, 0x00, 0x04}
//	"07-e0-06"       -> {0x07, 0xe0, 0x06}
//	"01-00-00-00-fe" -> {0x01, 0x00, 0x00, 0x00, 0xfe}
func ParseActionCode(s string) (ActionCode, error) {
	return parseActionCodeString(s)
}

func (a *ActionCode) UnmarshalYAML(node *yaml.Node) error {
	ac, err := parseActionCodeString(node.Value)
	*a = ac
	return err
}

func parseActionCodeString(s string) (ActionCode, error) {
	var ac ActionCode

	s = strings.TrimSpace(s)
	if s == "" {
		return ac, fmt.Errorf("empty action code")
	}

	parts := strings.Split(s, "-")
	if len(parts) < minActionCodeLen {
		return ac, fmt.Errorf("action code too short")
	}
	if len(parts) > maxActionCodeLen {
		return ac, fmt.Errorf("action code too long")
	}

	for i, p := range parts {
		p = strings.TrimSpace(p)

		if p == "" {
			return ac, fmt.Errorf("empty byte in action code")
		}

		v, err := strconv.ParseUint(p, 16, 8)
		if err != nil {
			return ac, fmt.Errorf("invalid byte %q", p)
		}

		ac[i] = byte(v)
	}

	return ac, nil
}

// ParseActionCodeBytes verifies the length of a byte slice,
// then removes right side padding and coerces the bytes
// into an ActionCode type.
func ParseActionCodeBytes(resp []byte) (ActionCode, error) {
	if len(resp) < minActionCodeLen {
		return ActionCode{}, fmt.Errorf("action code too short: got %d bytes, need at least %d", len(resp), minActionCodeLen)
	}

	// Find last non-zero byte
	end := len(resp)
	for end > 0 && resp[end-1] == 0x00 {
		end--
	}
	if end < minActionCodeLen {
		end = minActionCodeLen
	}

	if end > maxActionCodeLen {
		return ActionCode{}, fmt.Errorf("action code too long: got %d bytes, maximum %d", end, maxActionCodeLen)
	}
	var out ActionCode
	copy(out[:], resp[:end])
	return out, nil
}
