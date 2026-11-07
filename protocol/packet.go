package protocol

import (
	"fmt"

	"github.com/ColinEge/8bd-ultimate-api/keydefs"
)

// Packet size = 33 bytes
// [0]  response marker
// [1]  response marker
// [2..31] payload
// [32] continuation flag

type RemappedKey struct {
	PhysicalKey keydefs.KeyID
	Page        byte
}

func ParseRemappedKeysPackets(packets [][]byte) ([]RemappedKey, error) {
	var result []RemappedKey
	for _, p := range packets {
		if len(p) != ReportSize {
			return nil, fmt.Errorf("invalid packet length %d", len(p))
		}
		if p[0] != byte(DirIn) || p[1] != byte(CommandReadRemapedKeys) {
			return nil, fmt.Errorf("unexpected mapped-keys response header")
		}
		payload := p[2:]
		// 30 bytes usable for kv pairs
		for i := 0; i < 30; i += 2 {
			key := payload[i]
			page := payload[i+1]
			// firmware may pad empty entries
			if key == 0 && page == 0 {
				continue
			}
			result = append(result, RemappedKey{
				PhysicalKey: keydefs.KeyID(key),
				Page:        page,
			})
		}
	}
	return result, nil
}

// PadRight adds 0x00 bytes to the right of a slice until it meets
// ReportSize
func PadRight(bytes *[]byte) {
	if bytes == nil {
		return
	}
	if len(*bytes) >= ReportSize {
		return
	}
	if len(*bytes) < ReportSize {
		*bytes = append(*bytes, make([]byte, ReportSize-len(*bytes))...)
	}
}
