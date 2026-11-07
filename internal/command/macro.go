package command

import "github.com/ColinEge/8bd-ultimate-api/protocol"

// RecordMacro
type RecordMacro struct {
	ProfileName string
}

func (c *RecordMacro) Name() string {
	return "RecordMacro"
}

func (c *RecordMacro) Build() [][]byte {
	const placeholder = 0x00 // TODO: Update placeholder
	payload := protocol.UTF16BEEncode(c.ProfileName)
	return protocol.BuildText(protocol.Header(protocol.DirOut, protocol.CommandWriteProfile, placeholder), payload...)
}

func (c RecordMacro) ReadStrategy() ReadStrategy {
	return ReadText(1)
}

// packetToHID returns the HID as a full hex number including the 07 00 prefix.
// For example, A -> 0x070004
func packetToHID(packet []byte) (uint32, bool) {
	if len(packet) < 3 {
		return 0, false // invalid packet
	}

	bitmap := packet[2:] // skip HID page prefix

	for byteIndex, b := range bitmap {
		for bitPos := 0; bitPos < 8; bitPos++ {
			if b&(1<<bitPos) != 0 {
				bitIndex := byteIndex*8 + bitPos
				hid := uint32(bitIndex + 4)
				return hid, true
			}
		}
	}

	// Release event
	return 0x070000, false
}
