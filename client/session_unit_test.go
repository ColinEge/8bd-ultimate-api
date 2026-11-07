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

func newFixture(t *testing.T) *keyboard.KeyboardFixture {
	t.Helper()
	f := &keyboard.KeyboardFixture{}
	if err := f.Reset(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close fixture: %v", err)
		}
	})
	return f
}

func bindingPacket() []byte {
	packet := make([]byte, 33)
	copy(packet, []byte{0x52, 0xfa, 0x03, 0x0c, 0x00, 0xaa, 0x09, 0x71, 0x04, 0x07, 0x00, 0x05})
	return packet
}

func TestBindAction(t *testing.T) {
	writeErr := errors.New("write failed")
	readErr := errors.New("read failed")
	paddedOK := make([]byte, 33)
	copy(paddedOK, []byte{0x54, 0xe4, 0x08})
	oversized := make([]byte, 34)
	copy(oversized, paddedOK)
	tests := []struct {
		name     string
		response []byte
		writeErr error
		readErr  error
		wantErr  string
		cause    error
		reads    int
	}{
		{name: "ok", response: []byte{0x54, 0xe4, 0x08}, reads: 1},
		{name: "padded ok", response: paddedOK, reads: 1},
		{name: "rejected", response: []byte{0x54, 0xe4, 0x09}, wantErr: "device responded bad", reads: 1},
		{name: "ready is not success", response: []byte{0x54, 0x8a, 0x07, 0x01}, wantErr: "device responded ready", reads: 1},
		{name: "empty", wantErr: "unexpected status payload", reads: 1},
		{name: "truncated", response: []byte{0x54, 0xe4}, wantErr: "unexpected status payload", reads: 1},
		{name: "wrong prefix", response: []byte{0x52, 0xe4, 0x08}, wantErr: "unexpected status payload", reads: 1},
		{name: "unknown status", response: []byte{0x54, 0xe4, 0xff}, wantErr: "unexpected status payload", reads: 1},
		{name: "oversized", response: oversized, wantErr: "status packet too long", reads: 1},
		{name: "write failure", writeErr: writeErr, wantErr: "BindAction write failed", cause: writeErr},
		{name: "read failure", readErr: readErr, wantErr: "BindAction read failed", cause: readErr, reads: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.Exchanges = []keyboard.Exchange{{
				Write: bindingPacket(), Response: tt.response, WriteErr: tt.writeErr, ReadErr: tt.readErr,
			}}
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(s client.Keyboard) error {
				info, err := s.Info(context.Background())
				if err != nil || info.Device.ID != f.Device.ID || info.Device.Product != f.Device.Product {
					t.Errorf("incorrect keyboard metadata: %+v, %v", info, err)
				}
				return s.Bind(context.Background(), keydefs.KeyID(0x04), keydefs.ActionCode{0x07, 0x00, 0x05})
			})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
			}
			if tt.cause != nil && !errors.Is(err, tt.cause) {
				t.Errorf("lost error identity: %v", err)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
			if len(f.Writes) != 1 || f.Reads != tt.reads || f.SessionCloses != 1 {
				t.Fatalf("writes=%d reads=%d closes=%d", len(f.Writes), f.Reads, f.SessionCloses)
			}
		})
	}
}

func TestWithSession(t *testing.T) {
	callbackErr := errors.New("callback failed")
	closeErr := errors.New("close failed")
	connectErr := errors.New("connect failed")
	tests := []struct {
		name        string
		callbackErr error
		closeErr    error
		connectErr  error
		panicValue  interface{}
		wantText    string
	}{
		{name: "success"},
		{name: "callback error", callbackErr: callbackErr},
		{name: "close error", closeErr: closeErr},
		{name: "both errors", callbackErr: callbackErr, closeErr: closeErr},
		{name: "connect error", connectErr: connectErr},
		{name: "string panic", panicValue: "boom", wantText: "panic in session lambda: boom"},
		{name: "error panic", panicValue: callbackErr, wantText: "panic in session lambda"},
		{name: "panic and close error", panicValue: callbackErr, closeErr: closeErr, wantText: "panic in session lambda"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.SessionCloseErr = tt.closeErr
			f.OpenErr = tt.connectErr
			calls := 0
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(s client.Keyboard) error {
				calls++
				if tt.panicValue != nil {
					panic(tt.panicValue)
				}
				return tt.callbackErr
			})
			for _, cause := range []error{tt.callbackErr, tt.closeErr, tt.connectErr} {
				if cause != nil && !errors.Is(err, cause) {
					t.Errorf("got %v, lost error %v", err, cause)
				}
			}
			if cause, ok := tt.panicValue.(error); ok && !errors.Is(err, cause) {
				t.Errorf("lost panic error identity: %v", err)
			}
			if tt.wantText != "" && (err == nil || !strings.Contains(err.Error(), tt.wantText)) {
				t.Errorf("got %v, want %q", err, tt.wantText)
			}
			if tt.name == "success" && err != nil {
				t.Fatal(err)
			}
			wantCalls := 1
			if tt.connectErr != nil {
				wantCalls = 0
			}
			if calls != wantCalls || f.SessionCloses != wantCalls {
				t.Fatalf("callback calls=%d closes=%d, want %d", calls, f.SessionCloses, wantCalls)
			}
		})
	}
}
