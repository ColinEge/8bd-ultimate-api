package protocol

// BuildText constructs HID packets from a header and payload
// according to the packet framing observed during reverse engineering.
//
// header contains the command prefix (direction + group + optional params).
// Typical commands start with:
//
//	byte 0: direction (52 = out to keyboard, 54 = in from keyboard)
//	byte 1: command group
//
// Example:
//
//	52 80 -> out, readProfile
//	54 70 -> in, writeProfile
//
// Most commands fit within a single packet. If the payload is too large,
// the device splits it across two packets using the following fields:
//
//	byte 2: payload length in this packet
//	byte 3: remaining packet indicator
//
//	00 = this is the final packet
//	01 = could mean additional packet follows or a bool for keep reading
//	02 = 02 is theoretical, it would only exist if byte 3 is a count, not a bool (needs testing)
//
// No commands have been observed using more than two packets.
//
// Example: setting a two-packet profile name
//
// Packet 1:
//
//	52 70 1d 01 00 31 00 32 00 33 00 34 00 35 00 36 00 37 00 38 00 39 00 30 00 31 00 32 00 33 00 34 00
//
//	1d -> 29 bytes reported payload
//	01 -> one additional packet follows
//
// Packet 2:
//
//	52 70 03 00 35 00 36 00 00 00 00
//
//	03 -> 3 bytes of payload
//	00 -> final packet
//
// Packets are padded to the HID report size (33 bytes).
func BuildText(header []byte, payload ...byte) [][]byte {
	maxPayloadSize := ReportSize - len(header) - 2

	var packets [][]byte

	if payload == nil {
		if len(header) < ReportSize {
			header = append(header, make([]byte, ReportSize-len(header))...)
		}
		return [][]byte{header}
	}

	for offset := 0; offset < len(payload); offset += maxPayloadSize {
		end := offset + maxPayloadSize
		if end > len(payload) {
			end = len(payload)
		}

		chunk := payload[offset:end]

		pkt := append([]byte{}, header...)

		// remaining packets indicator
		remaining := byte(0)
		if end < len(payload) {
			remaining = 1
		}

		pkt = append(pkt, byte(len(chunk)), remaining)
		pkt = append(pkt, chunk...)

		if len(pkt) < ReportSize {
			pkt = append(pkt, make([]byte, ReportSize-len(pkt))...)
		}

		packets = append(packets, pkt)
	}

	return packets
}

// ReadText combines multiple packets into a single UTF16BE encoded byte slice
// paramOffset shifts the <len> <rem> lookup to the right or left,
// use when you've added params to the command
func ReadText(packets [][]byte, paramOffset int) []byte {
	var full []byte
	for _, pkt := range packets {
		payloadStart := 4 + paramOffset
		payloadLen := int(pkt[2+paramOffset])
		if payloadStart+payloadLen > len(pkt) {
			payloadLen = len(pkt) - payloadStart
		}
		full = append(full, pkt[payloadStart:payloadStart+payloadLen]...)
	}
	return full
}
