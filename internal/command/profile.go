package command

import (
	"fmt"
	"unicode/utf16"

	"github.com/ColinEge/8bd-ultimate-api/protocol"
)

type ProfileResponse struct {
	Payload packets
}

func (r *ProfileResponse) Decode() (string, error) {
	if len(r.Payload) == 0 || len(r.Payload) > 2 {
		return "", fmt.Errorf("invalid profile packet count %d", len(r.Payload))
	}
	var data []byte
	for i, p := range r.Payload {
		if len(p) < 4 || len(p) > protocol.ReportSize ||
			p[0] != byte(protocol.DirIn) || p[1] != byte(protocol.CommandReadProfile) ||
			int(p[2]) > protocol.ReportSize-4 || 4+int(p[2]) > len(p) ||
			(i < len(r.Payload)-1 && p[3] != 1) || (i == len(r.Payload)-1 && p[3] != 0) {
			return "", fmt.Errorf("invalid profile text packet %d", i)
		}
		data = append(data, p[4:4+int(p[2])]...)
	}
	if len(data)%2 != 0 {
		return "", fmt.Errorf("odd UTF-16 profile length")
	}
	for i := 0; i < len(data); i += 2 {
		unit := uint16(data[i])<<8 | uint16(data[i+1])
		if unit == 0 {
			return "", fmt.Errorf("NUL in profile title")
		}
		if utf16.IsSurrogate(rune(unit)) {
			if unit > 0xdbff || i+3 >= len(data) {
				return "", fmt.Errorf("unpaired UTF-16 surrogate")
			}
			next := uint16(data[i+2])<<8 | uint16(data[i+3])
			if next < 0xdc00 || next > 0xdfff {
				return "", fmt.Errorf("unpaired UTF-16 surrogate")
			}
			i += 2
		}
	}
	return protocol.UTF16BEDecode(data), nil
}

func (r *ProfileResponse) SetPayload(p packets) {
	r.Payload = p
}

// SetProfile
type SetProfile struct {
	ProfileName string
}

func (c *SetProfile) Name() string {
	return "SetProfile"
}

func (c *SetProfile) Build() packets {
	payload := protocol.UTF16BEEncode(c.ProfileName)
	return protocol.BuildText(protocol.Header(protocol.DirOut, protocol.CommandWriteProfile), payload...)
}

func (c *SetProfile) ReadStrategy() ReadStrategy {
	return ReadSingle()
}

// GetProfile
type GetProfile struct {
}

func (c *GetProfile) Name() string {
	return "GetProfile"
}

func (c *GetProfile) Build() packets {
	return protocol.BuildText(protocol.Header(protocol.DirOut, protocol.CommandReadProfile))
}

func (c *GetProfile) ReadStrategy() ReadStrategy {
	return ReadText(0)
}
