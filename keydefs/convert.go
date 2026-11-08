package keydefs

// byteSwapHWKeyBreakpoint is the highest PhysicalKey value that maps
// directly to a standard keyboard usage.
//
// PhysicalKey values above this threshold represent modifier or special
// keys. The firmware encodes these differently by placing the usage ID
// in byte 2 instead of byte 3.
const byteSwapHWKeyBreakpoint = 0x63

// byteSwapUsageBreakpoint is the first keyboard usage ID that uses the
// swapped encoding scheme when converted to a PhysicalKey.
//
// In the HID Keyboard/Keypad page (0x07), usages >= 0xE0 correspond to
// modifier keys. The firmware stores these keys in the PhysicalKey range
// using an offset and byte swap.
const byteSwapUsageBreakpoint = 0xe0

// conversionOffset is applied when converting between PhysicalKey values
// and modifier usages. The firmware compresses the modifier usage range
// by subtracting this offset when storing them as PhysicalKey values.
const conversionOffset = 0x7c

// hidPageKeyboard is the HID Usage Page for Keyboard/Keypad devices.
// Only actions on this page have a corresponding PhysicalKey.
const hidPageKeyboard = 0x07 // https://www.usb.org/sites/default/files/hut1_5.pdf#page=90

// ActionToPhysicalKey converts a keyboard ActionCode into the firmware's
// PhysicalKey identifier.
//
// This conversion only applies to keyboard actions on HID page 0x07
// that follow the standard 3-byte keyboard encoding:
//
//	[page]-[mod]-[key]
//
// For these actions:
//
//	Standard keys
//	  07-00-XX  → PhysicalKey = XX
//
//	Modifier keys
//	  07-E0-00  → PhysicalKey = E0 - conversionOffset
//
// Ordinary physical IDs end at 0x63; 0x64..0x6b encode modifiers,
// and 0xe8..0xe9 encode Super B/A defaults (07-E8-00 and 07-E9-00).
// This is firmware addressing, not HID usage-to-key identity. In particular,
// F13..F24, disabled actions, and chords have no conversion here.
// Returns (0, false) for actions outside these encodings.
func ActionToPhysicalKey(action ActionCode) (KeyID, bool) {
	if !action.IsKeyboardUsage() {
		return 0, false
	}

	byte1 := action[1]
	byte2 := action[2]

	// Standard key (07-00-XX)
	if byte1 == 0x00 && byte2 >= 4 && byte2 <= byteSwapHWKeyBreakpoint {
		return KeyID(byte2), true
	}

	// Super keys use their firmware usage byte directly as the physical ID.
	if (byte1 == 0xe8 || byte1 == 0xe9) && byte2 == 0 {
		return KeyID(byte1), true
	}

	// Modifier key (07-E0-00 and above)
	if byte1 >= byteSwapUsageBreakpoint && byte1 <= 0xe7 && byte2 == 0 {
		return KeyID(byte1 - conversionOffset), true
	}

	// Other reserved encodings
	return 0, false
}

func (a ActionCode) IsKeyboardUsage() bool {
	return a.Len() == 3 && a[0] == hidPageKeyboard
}

// PhysicalKeyToAction converts a firmware PhysicalKey identifier into
// the corresponding keyboard ActionCode.
//
// The firmware uses two encoding schemes:
//
//	Standard keys
//	  PhysicalKey <= byteSwapHWKeyBreakpoint
//	  → [07, 00, PhysicalKey]
//
//	Modifier/special keys
//	  PhysicalKey > byteSwapHWKeyBreakpoint
//	  → [07, PhysicalKey + conversionOffset, 00]
//
// Super B/A instead use physical IDs 0xe8/0xe9 and actions [07, ID, 00].
// Valid firmware IDs are 0x04..0x6b and 0xe8..0xe9. Other IDs return zero;
// callers performing I/O must validate the ID first.
func PhysicalKeyToAction(key byte) ActionCode {
	if key == 0xe8 || key == 0xe9 {
		return ActionCode{hidPageKeyboard, key, 0x00}
	}
	if key < 4 || key > 0x6b {
		return ActionCode{}
	}
	if key <= byteSwapHWKeyBreakpoint {
		return ActionCode{hidPageKeyboard, 0x00, key}
	}
	return ActionCode{hidPageKeyboard, key + conversionOffset, 0x00}
}
