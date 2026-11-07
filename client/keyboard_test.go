package client_test

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ColinEge/8bd-kbd-api/client"
	"github.com/ColinEge/8bd-kbd-api/test/fixtures/keyboard"
	"github.com/ColinEge/8bd-kbd-api/transport"
)

func bindRequest(key client.KeyID, action client.Action) []byte {
	packet := bindingPacket()
	packet[8] = byte(key)
	copy(packet[9:], action[:])
	return packet
}

func TestKeyboardActions(t *testing.T) {
	copyAction, err := client.Chord(client.LeftCtrl, client.ActionC())
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []client.Action{
		client.ActionB(), copyAction, client.ActionF13(), client.ActionDisabled(),
		client.ActionVolumeUp(), client.ActionVolumeDown(), client.ActionMute(),
	} {
		t.Run(action.String(), func(t *testing.T) {
			f := newFixture(t)
			f.Exchanges = []keyboard.Exchange{{Write: bindRequest(client.KeyA, action), Response: []byte{0x54, 0xe4, 8}}}
			if err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
				if action == client.ActionDisabled() {
					return k.Disable(context.Background(), client.KeyA)
				}
				return k.Bind(context.Background(), client.KeyA, action)
			}); err != nil {
				t.Fatal(err)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestKeyboardRejectsInvalidChangesWithoutIO(t *testing.T) {
	for _, test := range []struct {
		name    string
		changes []client.BindingChange
		want    error
	}{
		{"invalid key", []client.BindingChange{{Key: 0, Action: client.ActionB()}}, client.ErrInvalidKey},
		{"zero action", []client.BindingChange{{Key: client.KeyA}}, client.ErrInvalidAction},
		{"unsupported", []client.BindingChange{{Key: client.KeyA, Action: client.Action{0xff, 1, 2}}}, client.ErrUnsupported},
		{"invalid second", []client.BindingChange{{Key: client.KeyA, Action: client.ActionB()}, {Key: 255, Action: client.ActionB()}}, client.ErrInvalidKey},
		{"duplicate", []client.BindingChange{{Key: client.KeyA, Action: client.ActionB()}, {Key: client.KeyA, Action: client.ActionC()}}, client.ErrInvalidKey},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newFixture(t)
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
				return k.ApplyBindings(context.Background(), test.changes)
			})
			if !errors.Is(err, test.want) || len(f.Writes) != 0 {
				t.Fatalf("error=%v writes=%d", err, len(f.Writes))
			}
		})
	}
	for _, text := range []string{"", "07", "07-00", "07-zz-04"} {
		if _, err := client.ParseAction(text); !errors.Is(err, client.ErrInvalidAction) {
			t.Fatalf("parse %q: %v", text, err)
		}
	}
	for _, action := range []client.Action{client.ActionDisabled(), {1, 1, 0}, {7, 0xe0, 6}} {
		if _, err := client.Chord(client.LeftCtrl, action); !errors.Is(err, client.ErrInvalidAction) {
			t.Fatalf("chord %s: %v", action, err)
		}
	}
}

func TestKeyboardBindingStates(t *testing.T) {
	for _, test := range []struct {
		action    client.Action
		state     client.BindingState
		effective client.Action
	}{
		{client.Action{}, client.BindingDefault, client.ActionA()},
		{client.ActionDisabled(), client.BindingDisabled, client.ActionDisabled()},
		{client.ActionB(), client.BindingAction, client.ActionB()},
	} {
		f := newFixture(t)
		f.Exchanges = []keyboard.Exchange{{Write: actionRequest(4), Response: actionReply(4, test.action)}}
		err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
			binding, err := k.Binding(context.Background(), client.KeyA)
			if err == nil && (binding.Key != client.KeyA || binding.Action != test.action || binding.State != test.state || binding.Effective != test.effective) {
				t.Fatalf("unexpected binding: %+v", binding)
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := f.AssertComplete(); err != nil {
			t.Fatal(err)
		}
	}
	f := newFixture(t)
	f.Exchanges = []keyboard.Exchange{{Write: actionRequest(4), Response: actionReply(5, client.ActionB())}}
	err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
		_, err := k.Binding(context.Background(), client.KeyA)
		return err
	})
	if !errors.Is(err, client.ErrMalformedResponse) {
		t.Fatalf("wrong response key: %v", err)
	}
}

func TestKeyboardApplyBindings(t *testing.T) {
	for _, fail := range []bool{false, true} {
		f := newFixture(t)
		cause := errors.New("write failed")
		changes := []client.BindingChange{{Key: client.KeyA, Action: client.ActionB()}, {Key: client.KeyB, Action: client.ActionC()}}
		f.Exchanges = []keyboard.Exchange{
			{Write: bindRequest(changes[0].Key, changes[0].Action), Response: []byte{0x54, 0xe4, 8}},
			{Write: bindRequest(changes[1].Key, changes[1].Action), Response: []byte{0x54, 0xe4, 8}},
		}
		if fail {
			f.Exchanges[1].WriteErr = cause
		}
		err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
			return k.ApplyBindings(context.Background(), changes)
		})
		if fail {
			var applyErr *client.ApplyError
			if !errors.Is(err, cause) || !errors.As(err, &applyErr) ||
				!reflect.DeepEqual(applyErr.Completed, changes[:1]) || applyErr.Failed != changes[1] {
				t.Fatalf("partial application: %v", err)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(changes, []client.BindingChange{{Key: client.KeyA, Action: client.ActionB()}, {Key: client.KeyB, Action: client.ActionC()}}) {
			t.Fatal("input changed")
		}
		if err := f.AssertComplete(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestKeyboardBindings(t *testing.T) {
	request := make([]byte, 33)
	copy(request, []byte{0x52, 0x81})
	for _, test := range []struct {
		name      string
		reply     []byte
		wantError bool
	}{
		{"empty", append([]byte{0x54, 0x81}, make([]byte, 31)...), false},
		{"mapped", append([]byte{0x54, 0x81, 4, 7}, make([]byte, 29)...), false},
		{"short", []byte{0x54}, true},
		{"wrong header", make([]byte, 33), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newFixture(t)
			f.Exchanges = []keyboard.Exchange{{Write: request, Response: test.reply}}
			if test.name == "mapped" {
				f.Exchanges = append(f.Exchanges, keyboard.Exchange{Write: actionRequest(4), Response: actionReply(4, client.ActionB())})
			}
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
				bindings, err := k.Bindings(context.Background())
				if !test.wantError && err == nil {
					want := []client.Binding{}
					if test.name == "mapped" {
						want = append(want, client.Binding{Key: client.KeyA, State: client.BindingAction, Action: client.ActionB(), Effective: client.ActionB()})
					}
					if !reflect.DeepEqual(bindings, want) {
						t.Fatalf("bindings=%+v want=%+v", bindings, want)
					}
				}
				return err
			})
			if test.wantError {
				if !errors.Is(err, client.ErrMalformedResponse) {
					t.Fatalf("malformed response: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestKeyboardCatalogAndLifecycle(t *testing.T) {
	f := newFixture(t)
	k, err := f.Keyboard.Connect(context.Background(), f.Device.ID)
	if err != nil {
		t.Fatal(err)
	}
	info, err := k.Info(context.Background())
	if err != nil || info.Device != f.Device {
		t.Fatalf("metadata=%+v error=%v", info, err)
	}
	if _, err := k.FirmwareVersion(context.Background()); !errors.Is(err, client.ErrUnsupported) {
		t.Fatalf("firmware: %v", err)
	}
	keys, actions := k.Keys(), k.Actions()
	if len(keys) == 0 || len(actions) == 0 {
		t.Fatal("empty built-in catalog")
	}
	seen := make(map[client.KeyID]bool)
	for _, key := range keys {
		if seen[key.ID] {
			t.Fatalf("duplicate catalog key %s", key.ID)
		}
		seen[key.ID] = true
		f.Exchanges = append(f.Exchanges, keyboard.Exchange{Write: actionRequest(byte(key.ID)), Response: actionReply(byte(key.ID), client.Action{})})
		binding, err := k.Binding(context.Background(), key.ID)
		if err != nil || binding.Effective != key.Default {
			t.Fatalf("catalog default %s: %+v, %v", key.ID, binding, err)
		}
	}
	if err := f.AssertComplete(); err != nil {
		t.Fatal(err)
	}
	first := actions[0]
	actions[0].Label = "changed"
	if k.Actions()[0] != first {
		t.Fatal("catalog is mutable")
	}
	if err := f.Keyboard.Close(); err != nil {
		t.Fatal(err)
	}
	if err := k.Close(); err != nil {
		t.Fatal(err)
	}
	if f.SessionCloses != 1 || f.ProviderCloses != 1 {
		t.Fatalf("closes: %d, %d", f.SessionCloses, f.ProviderCloses)
	}
	if _, err := k.Info(context.Background()); !errors.Is(err, client.ErrDeviceClosed) {
		t.Fatalf("closed info: %v", err)
	}
}

type apiProvider struct {
	transport transport.Transport
}

func (p *apiProvider) ListDevices(ctx context.Context) ([]client.DeviceInfo, error) {
	return nil, ctx.Err()
}
func (p *apiProvider) OpenDevice(ctx context.Context, id client.DeviceID) (*client.OpenedDevice, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &client.OpenedDevice{Transport: p.transport, Info: client.DeviceInfo{ID: id, VendorID: 0x2dc8, Product: client.ProductTypeKeyboard}}, nil
}
func (p *apiProvider) Close() error { return nil }

func TestKeyboardCancellationAndSerialization(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var writes atomic.Int32
	c := client.New(client.WithDeviceProvider(&apiProvider{transport: &transport.Mock{
		WriteFunc: func(ctx context.Context, data []byte) error {
			writes.Add(1)
			close(started)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-release:
				return nil
			}
		},
		ReadFunc: func(context.Context) ([]byte, error) { return []byte{0x54, 0xe4, 8}, nil },
	}}))
	defer c.Close()
	k, err := c.Connect(context.Background(), "001-003")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- k.Bind(ctx, client.KeyA, client.ActionB()) }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("write did not start")
	}
	waitCtx, cancelWait := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelWait()
	if err := k.Bind(waitCtx, client.KeyB, client.ActionC()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting command: %v", err)
	}
	if writes.Load() != 1 {
		t.Fatal("commands interleaved")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled write: %v", err)
	}
}

func TestKeyboardWriteTimeout(t *testing.T) {
	c := client.New(client.WithTimeout(20*time.Millisecond), client.WithDeviceProvider(&apiProvider{transport: &transport.Mock{
		WriteFunc: func(ctx context.Context, data []byte) error { <-ctx.Done(); return ctx.Err() },
	}}))
	defer c.Close()
	err := c.WithKeyboard(context.Background(), "001-003", func(k client.Keyboard) error { return k.Bind(context.Background(), client.KeyA, client.ActionB()) })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("write timeout: %v", err)
	}
}

func TestKeyboardReadTimeout(t *testing.T) {
	c := client.New(client.WithTimeout(20*time.Millisecond), client.WithDeviceProvider(&apiProvider{transport: &transport.Mock{
		WriteFunc: func(context.Context, []byte) error { return nil },
		ReadFunc:  func(ctx context.Context) ([]byte, error) { <-ctx.Done(); return nil, ctx.Err() },
	}}))
	defer c.Close()
	err := c.WithKeyboard(context.Background(), "001-003", func(k client.Keyboard) error { _, err := k.Binding(context.Background(), client.KeyA); return err })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("read timeout: %v", err)
	}
}

func TestKeyboardDiscoveryAndReconnection(t *testing.T) {
	f := newFixture(t)
	for _, connected := range []bool{true, false, true} {
		f.Connected = connected
		devices, err := f.Keyboard.Scan(context.Background())
		if err != nil || (len(devices) == 1) != connected {
			t.Fatalf("connected=%t devices=%+v error=%v", connected, devices, err)
		}
		if !connected {
			if _, err := f.Keyboard.Connect(context.Background(), f.Device.ID); !errors.Is(err, client.ErrDeviceNotFound) {
				t.Fatalf("disconnected open: %v", err)
			}
		} else {
			k, err := f.Keyboard.Connect(context.Background(), f.Device.ID)
			if err != nil {
				t.Fatal(err)
			}
			if err := k.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := f.Keyboard.Scan(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled scan: %v", err)
	}
	if _, err := f.Keyboard.Connect(ctx, f.Device.ID); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled connect: %v", err)
	}
}
