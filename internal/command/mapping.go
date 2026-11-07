package command

import (
	"fmt"

	"github.com/ColinEge/8bd-ultimate-api/keydefs"
	"github.com/ColinEge/8bd-ultimate-api/protocol"
)

// RemappedKeysResponse
type RemappedKeysResponse struct {
	Payload packets
}

func (r *RemappedKeysResponse) Decode() ([]protocol.RemappedKey, error) {
	rk, err := protocol.ParseRemappedKeysPackets(r.Payload)
	if err != nil {
		return nil, err
	}
	return rk, nil
}

func (r *RemappedKeysResponse) SetPayload(p packets) {
	r.Payload = p
}

// GetRemappedKeys
type GetRemappedKeys struct {
}

func (c *GetRemappedKeys) Name() string {
	return "GetRemappedKeys"
}

func (c *GetRemappedKeys) Build() packets {
	return protocol.BuildText(protocol.Header(protocol.DirOut, protocol.CommandReadRemapedKeys))
}

func (c GetRemappedKeys) ReadStrategy() ReadStrategy {
	return ReadContinuation(32)
}

// KeyActionResponse decodes one key-action packet.
type KeyActionResponse struct {
	Payload    []byte
	payloadErr error
}

func (r *KeyActionResponse) Decode() (keydefs.ActionCode, error) {
	if r.payloadErr != nil {
		return keydefs.ActionCode{}, r.payloadErr
	}
	if len(r.Payload) < 6 || len(r.Payload) > protocol.ReportSize {
		return keydefs.ActionCode{}, fmt.Errorf("invalid key action packet length %d: expected 6 to %d bytes", len(r.Payload), protocol.ReportSize)
	}
	if r.Payload[0] != byte(protocol.DirIn) ||
		r.Payload[1] != byte(protocol.CommandReadKeyAction) {
		return keydefs.ActionCode{}, fmt.Errorf("unexpected key action header %x", r.Payload[:2])
	}
	return keydefs.ParseActionCodeBytes(r.Payload[3:])
}

func (r *KeyActionResponse) SetPayload(p packets) {
	r.Payload = nil
	r.payloadErr = nil
	if len(p) != 1 {
		r.payloadErr = fmt.Errorf("unexpected key action packet count %d: expected 1", len(p))
		return
	}
	r.Payload = p[0]
}

// GetKeyAction requests the action assigned to a physical key.
type GetKeyAction struct {
	PhysicalKey keydefs.KeyID
}

func (c *GetKeyAction) Name() string {
	return "GetKeyAction"
}

func (c *GetKeyAction) Build() packets {
	var pkt []byte
	pkt = append(pkt, protocol.Header(protocol.DirOut, protocol.CommandReadKeyAction)...)
	pkt = append(pkt, byte(c.PhysicalKey))
	protocol.PadRight(&pkt)
	return [][]byte{pkt}
}

func (c GetKeyAction) ReadStrategy() ReadStrategy {
	return ReadSingle()
}
