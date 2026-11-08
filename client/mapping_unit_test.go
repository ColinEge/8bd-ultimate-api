package client_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
	"github.com/ColinEge/8bd-ultimate-api/test/fixtures/keyboard"
)

func actionRequest(key byte) []byte {
	packet := make([]byte, 33)
	copy(packet, []byte{0x52, 0x83, key})
	return packet
}

func readAction(ctx context.Context, keyboard client.Keyboard, key client.KeyID) (client.Action, bool, error) {
	binding, err := keyboard.Binding(ctx, key)
	return binding.Action, err == nil && binding.State != client.BindingDefault, err
}

func actionReply(key byte, action keydefs.ActionCode) []byte {
	packet := make([]byte, 33)
	copy(packet, []byte{0x54, 0x83, key})
	copy(packet[3:], action[:])
	return packet
}

func TestAction(t *testing.T) {
	writeErr := errors.New("write failure")
	readErr := errors.New("read failure")
	for _, tt := range []struct {
		name      string
		reply     []byte
		want      keydefs.ActionCode
		writeErr  error
		readErr   error
		errorText string
		cause     error
	}{
		{name: "minimal", reply: []byte{0x54, 0x83, 4, 7, 0, 5}, want: keydefs.ActionCode{7, 0, 5}},
		{name: "padded", reply: actionReply(4, keydefs.ActionCode{7, 0, 5}), want: keydefs.ActionCode{7, 0, 5}},
		{name: "unbound hardware response", reply: actionReply(4, keydefs.ActionCode{})},
		{name: "unbound minimal", reply: []byte{0x54, 0x83, 4, 0, 0, 0}},
		{name: "disabled is mapped", reply: actionReply(4, keydefs.ActionCode{7, 0, 0}), want: keydefs.ActionCode{7, 0, 0}},
		{name: "modifier", reply: actionReply(4, keydefs.ActionCode{7, 0xe0, 0}), want: keydefs.ActionCode{7, 0xe0, 0}},
		{name: "combination", reply: actionReply(4, keydefs.ActionCode{7, 0xe0, 6}), want: keydefs.ActionCode{7, 0xe0, 6}},
		{name: "six byte action", reply: actionReply(4, keydefs.ActionCode{1, 2, 3, 4, 5, 6}), want: keydefs.ActionCode{1, 2, 3, 4, 5, 6}},
		{name: "empty", errorText: "invalid key action packet length"},
		{name: "truncated header", reply: []byte{0x54}, errorText: "invalid key action packet length"},
		{name: "truncated action", reply: []byte{0x54, 0x83, 4, 7, 0}, errorText: "invalid key action packet length"},
		{name: "wrong header", reply: []byte{0x52, 0x83, 4, 7, 0, 5}, errorText: "unexpected key action header"},
		{name: "oversized packet", reply: make([]byte, 34), errorText: "invalid key action packet length"},
		{name: "oversized action", reply: []byte{0x54, 0x83, 4, 1, 2, 3, 4, 5, 6, 7}, errorText: "action code too long"},
		{name: "write failure", writeErr: writeErr, cause: writeErr, errorText: "GetKeyAction write failed"},
		{name: "read failure", readErr: readErr, cause: readErr, errorText: "GetKeyAction read failed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.Exchanges = []keyboard.Exchange{{Write: actionRequest(4), Response: tt.reply, WriteErr: tt.writeErr, ReadErr: tt.readErr}}
			var got keydefs.ActionCode
			var mapped bool
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(s client.Keyboard) error {
				var err error
				got, mapped, err = readAction(context.Background(), s, keydefs.KeyID(4))
				return err
			})
			if tt.errorText == "" {
				if err != nil || got != tt.want {
					t.Fatalf("got %v, %v; want %v", got, err, tt.want)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.errorText) || got != (keydefs.ActionCode{}) {
				t.Fatalf("got %v, %v; want zero action and %q", got, err, tt.errorText)
			}
			if tt.cause != nil && !errors.Is(err, tt.cause) {
				t.Fatalf("lost transport error: %v", err)
			}
			if wantMapped := tt.errorText == "" && tt.want != (keydefs.ActionCode{}); mapped != wantMapped {
				t.Fatalf("mapped=%t, want %t", mapped, wantMapped)
			}
			if f.SessionCloses != 1 {
				t.Fatalf("session closes=%d, want 1", f.SessionCloses)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResetToDefault(t *testing.T) {
	transportErr := errors.New("reset to default transport failure")
	for _, tt := range []struct {
		name      string
		key       byte
		action    keydefs.ActionCode
		reply     []byte
		writeErr  error
		readErr   error
		errorText string
	}{
		{name: "A", key: 4, action: keydefs.ActionCode{7, 0, 4}, reply: []byte{0x54, 0xe4, 8}},
		{name: "modifier", key: 0x66, action: keydefs.ActionCode{7, 0xe2, 0}, reply: []byte{0x54, 0xe4, 8}},
		{name: "Super B", key: 0xe8, reply: []byte{0x54, 0xe4, 8}},
		{name: "Super A", key: 0xe9, reply: []byte{0x54, 0xe4, 8}},
		{name: "Super rejected", key: 0xe8, reply: []byte{0x54, 0xe4, 9}, errorText: "device responded bad"},
		{name: "Super write failure", key: 0xe9, writeErr: transportErr, errorText: "BindAction write failed"},
		{name: "Super read failure", key: 0xe8, readErr: transportErr, errorText: "BindAction read failed"},
		{name: "rejected", key: 4, action: keydefs.ActionCode{7, 0, 4}, reply: []byte{0x54, 0xe4, 9}, errorText: "device responded bad"},
		{name: "malformed", key: 4, action: keydefs.ActionCode{7, 0, 4}, reply: []byte{0x54}, errorText: "unexpected status payload"},
		{name: "write failure", key: 4, action: keydefs.ActionCode{7, 0, 4}, writeErr: transportErr, errorText: "BindAction write failed"},
		{name: "read failure", key: 4, action: keydefs.ActionCode{7, 0, 4}, readErr: transportErr, errorText: "BindAction read failed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			packet := bindingPacket()
			packet[8] = tt.key
			copy(packet[9:], tt.action[:])
			f.Exchanges = []keyboard.Exchange{{Write: packet, Response: tt.reply, WriteErr: tt.writeErr, ReadErr: tt.readErr}}
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(s client.Keyboard) error {
				return s.ResetToDefault(context.Background(), keydefs.KeyID(tt.key))
			})
			if tt.errorText == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "reset key") || !strings.Contains(err.Error(), tt.errorText) {
				t.Fatalf("got %v, want reset to default error containing %q", err, tt.errorText)
			}
			if (tt.writeErr != nil || tt.readErr != nil) && !errors.Is(err, transportErr) {
				t.Fatalf("lost transport error: %v", err)
			}
			if f.SessionCloses != 1 {
				t.Fatalf("session closes=%d, want 1", f.SessionCloses)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
