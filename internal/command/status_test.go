package command_test

import (
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/internal/command"
)

func TestStatusResponse(t *testing.T) {
	ok := []byte{0x54, 0xe4, 0x08}
	for _, tt := range []struct {
		name    string
		packets [][]byte
	}{
		{name: "no packets"},
		{name: "too many packets", packets: [][]byte{ok, ok}},
		{name: "empty packet", packets: [][]byte{nil}},
		{name: "oversized packet", packets: [][]byte{make([]byte, 34)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &command.StatusResponse{}
			r.SetPayload([][]byte{ok})
			r.SetPayload(tt.packets)
			if status, err := r.DecodeStatus(); err == nil || status != "" {
				t.Fatalf("got status %q, error %v; want decoding failure", status, err)
			}
			if r.Decode() == command.StatusResponseOk {
				t.Fatal("invalid payload reused the previous successful status")
			}
			r.SetPayload([][]byte{ok})
			if status, err := r.DecodeStatus(); err != nil || status != command.StatusResponseOk {
				t.Fatalf("valid payload after error: status %q, error %v", status, err)
			}
		})
	}
}
