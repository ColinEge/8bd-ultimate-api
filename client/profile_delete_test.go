package client_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/test/fixtures/keyboard"
)

func TestDeleteProfile(t *testing.T) {
	transportErr := errors.New("USB failure")
	for _, tt := range []struct {
		name    string
		reply   []byte
		readErr error
		cause   error
	}{
		{"success", []byte{0x54, 0xe4, 8}, nil, nil},
		{"rejected", []byte{0x54, 0xe4, 9}, nil, client.ErrDeviceRejected},
		{"unknown", []byte{0x54, 0xe4, 7}, nil, client.ErrMalformedResponse},
		{"transport", nil, transportErr, transportErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			packet := make([]byte, 33)
			copy(packet, []byte{0x52, 0x70, 0, 0})
			f.Exchanges = []keyboard.Exchange{{Write: packet, Response: tt.reply, ReadErr: tt.readErr}}
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
				cancelled, cancel := context.WithCancel(context.Background())
				cancel()
				if err := k.DeleteProfile(cancelled); !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation: %v", err)
				}
				return k.DeleteProfile(context.Background())
			})
			if (tt.cause == nil && err != nil) || (tt.cause != nil && !errors.Is(err, tt.cause)) {
				t.Fatalf("error=%v, want=%v", err, tt.cause)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
			k, err := f.Keyboard.Connect(context.Background(), f.Device.ID)
			if err != nil {
				t.Fatal(err)
			}
			if err := k.Close(); err != nil {
				t.Fatal(err)
			}
			if err := k.DeleteProfile(context.Background()); !errors.Is(err, client.ErrDeviceClosed) {
				t.Fatalf("closed deletion: %v", err)
			}
		})
	}
}
