package command_test

import (
	"testing"

	"github.com/ColinEge/8bd-kbd-api/internal/command"
	"github.com/ColinEge/8bd-kbd-api/keydefs"
)

func TestKeyActionResponse(t *testing.T) {
	valid := [][]byte{{0x54, 0x83, 0x04, 7, 0, 5}}
	for _, tt := range []struct {
		name    string
		packets [][]byte
	}{
		{name: "no packets"},
		{name: "multiple packets", packets: append(valid, valid...)},
		{name: "nil packet", packets: [][]byte{nil}},
		{name: "short header", packets: [][]byte{{0x54}}},
		{name: "short action", packets: [][]byte{{0x54, 0x83, 4, 7, 0}}},
		{name: "wrong direction", packets: [][]byte{{0x52, 0x83, 4, 7, 0, 5}}},
		{name: "wrong command", packets: [][]byte{{0x54, 0x80, 4, 7, 0, 5}}},
		{name: "oversized packet", packets: [][]byte{make([]byte, 34)}},
		{name: "oversized action", packets: [][]byte{{0x54, 0x83, 4, 1, 2, 3, 4, 5, 6, 7}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &command.KeyActionResponse{}
			r.SetPayload(valid)
			r.SetPayload(tt.packets)
			if action, err := r.Decode(); err == nil || action != (keydefs.ActionCode{}) {
				t.Fatalf("got action %v, error %v; want explicit failure", action, err)
			}
			r.SetPayload(valid)
			if action, err := r.Decode(); err != nil || action != (keydefs.ActionCode{7, 0, 5}) {
				t.Fatalf("valid response after error: %v, %v", action, err)
			}
		})
	}
}
