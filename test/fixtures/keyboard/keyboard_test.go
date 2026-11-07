package keyboard

import (
	"context"
	"errors"
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
)

func TestResetClosesPreviousProvider(t *testing.T) {
	f := &KeyboardFixture{}
	if err := f.Reset(); err != nil {
		t.Fatal(err)
	}
	if err := f.RequireConnected(); err != nil {
		t.Fatal(err)
	}
	old := f.Keyboard
	closeErr := errors.New("provider close failure")
	f.ProviderCloseErr = closeErr
	if err := f.Reset(); !errors.Is(err, closeErr) {
		t.Fatalf("got %v, want %v", err, closeErr)
	}
	if f.Keyboard != old || f.ProviderCloses != 1 {
		t.Fatal("reset discarded a provider that failed to close")
	}
	f.ProviderCloseErr = nil
	if err := f.Reset(); err != nil {
		t.Fatal(err)
	}
	if f.Keyboard == old {
		t.Fatal("reset did not replace the old client")
	}
	if _, err := old.Scan(context.Background()); !errors.Is(err, client.ErrDeviceClosed) {
		t.Fatalf("closed client scan: got %v, want closed-device error", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if f.ProviderCloses != 1 {
		t.Fatalf("got %d closes, want 1", f.ProviderCloses)
	}
	if err := f.RequireConnected(); err == nil {
		t.Fatal("closed fixture reported connected")
	}
}

func TestScriptRejectsUnexpectedIO(t *testing.T) {
	for _, name := range []string{"unexpected write", "wrong packet", "missing read", "extra read"} {
		t.Run(name, func(t *testing.T) {
			f := &KeyboardFixture{Exchanges: []Exchange{{Write: []byte{1}, Response: []byte{2}}}}
			switch name {
			case "unexpected write":
				f.Exchanges = nil
				if err := f.write([]byte{1}); err == nil {
					t.Fatal("unexpected write succeeded")
				}
			case "wrong packet":
				if err := f.write([]byte{3}); err == nil {
					t.Fatal("wrong packet succeeded")
				}
			case "missing read":
				if err := f.write([]byte{1}); err != nil {
					t.Fatal(err)
				}
			case "extra read":
				if _, err := f.read(context.Background()); err == nil {
					t.Fatal("unexpected read succeeded")
				}
			}
			if err := f.AssertComplete(); err == nil {
				t.Fatal("invalid script reported complete")
			}
		})
	}
}

func TestFixtureRejectsDisconnectedDevice(t *testing.T) {
	f := &KeyboardFixture{}
	if err := f.Reset(); err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.Connected = false
	if devices, err := f.Keyboard.Scan(context.Background()); err != nil || len(devices) != 0 {
		t.Fatalf("scan: %v, %v", devices, err)
	}
	if _, err := f.Keyboard.Connect(context.Background(), f.Device.ID); !errors.Is(err, client.ErrDeviceNotFound) {
		t.Fatalf("connect: %v", err)
	}
}

func TestFixtureRejectsUseAfterSessionClose(t *testing.T) {
	f := &KeyboardFixture{}
	if err := f.Reset(); err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	s, err := f.Keyboard.Connect(context.Background(), f.Device.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Bind(context.Background(), keydefs.KeyID(4), keydefs.ActionCode{7, 0, 5}); !errors.Is(err, client.ErrDeviceClosed) {
		t.Fatalf("got %v, want closed-device error", err)
	}
}
