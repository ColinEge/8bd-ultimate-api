package protocol

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
)

// Build UTF16BE payload for profile names
func UTF16BEEncode(s string) []byte {
	runes := utf16.Encode([]rune(s))
	buf := new(bytes.Buffer)
	for _, r := range runes {
		_ = binary.Write(buf, binary.BigEndian, r)
	}
	return buf.Bytes()
}

// Build UTF16BE payload for profile names
func UTF16BEDecode(b []byte) string {
	if len(b)%2 != 0 {
		// Ignore trailing byte if odd length
		b = b[:len(b)-1]
	}

	runes := make([]uint16, len(b)/2)
	for i := 0; i < len(runes); i++ {
		runes[i] = binary.BigEndian.Uint16(b[i*2 : i*2+2])
	}

	return string(utf16.Decode(runes))
}

// TrimPadding returns s without trailing 0x00 bytes.
// Only trailing zeros are removed; interior zeros are preserved.
//
// Returns a subslice of the original slice, not a copy.
func TrimPadding(s []byte) []byte {
	i := len(s)
	for i > 0 && s[i-1] == 0x00 {
		i--
	}
	return s[:i]
}
