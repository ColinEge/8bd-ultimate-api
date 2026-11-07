package command

import (
	"fmt"

	"github.com/ColinEge/8bd-ultimate-api/protocol"
)

const (
	StatusResponseOk    = "ok"
	StatusResponseBad   = "bad"
	StatusResponseReady = "ready"
)

type StatusResponse struct {
	Payload    []byte
	payloadErr error
}

// DecodeStatus rejects malformed and unknown acknowledgments.
func (r *StatusResponse) DecodeStatus() (string, error) {
	if r.payloadErr != nil {
		return "", r.payloadErr
	}
	switch string(r.Payload) {
	case "\x54\xe4\x08":
		return StatusResponseOk, nil
	case "\x54\xe4\x09":
		return StatusResponseBad, nil
	case "\x54\x8a\x07\x01":
		return StatusResponseReady, nil
	default:
		return "", fmt.Errorf("unexpected status payload %x", r.Payload)
	}
}

// Decode preserves the string-only contract; use DecodeStatus to handle errors.
func (r *StatusResponse) Decode() string {
	status, err := r.DecodeStatus()
	if err != nil {
		return fmt.Sprintf("Unknown status %x", r.Payload)
	}
	return status
}

func (r *StatusResponse) SetPayload(p [][]byte) {
	r.Payload = nil
	r.payloadErr = nil
	const expectedPackets = 1
	if len(p) != expectedPackets {
		r.payloadErr = fmt.Errorf("unexpected status payload, got %d packets but expected %d", len(p), expectedPackets)
		return
	}
	if len(p[0]) > protocol.ReportSize {
		r.payloadErr = fmt.Errorf("status packet too long: got %d bytes", len(p[0]))
		return
	}
	r.Payload = protocol.TrimPadding(p[0])
}
