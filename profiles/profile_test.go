package profiles

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/test/fixtures/keyboard"
)

func newFixture(t *testing.T) *keyboard.KeyboardFixture {
	t.Helper()
	f := &keyboard.KeyboardFixture{}
	if err := f.Reset(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.AssertComplete(); err != nil {
			t.Error(err)
		}
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})
	return f
}

func bindingPacket() []byte {
	p := make([]byte, 33)
	copy(p, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 5})
	return p
}

func profileTextPackets(header byte, title string) [][]byte {
	var data []byte
	for _, unit := range utf16.Encode([]rune(title)) {
		data = append(data, byte(unit>>8), byte(unit))
	}
	var packets [][]byte
	for len(data) > 0 {
		n := min(29, len(data))
		p := make([]byte, 33)
		p[0], p[1], p[2] = 0x52, header, byte(n)
		if len(data) > n {
			p[3] = 1
		}
		copy(p[4:], data[:n])
		packets = append(packets, p)
		data = data[n:]
	}
	return packets
}

func TestKeyboardProfileName(t *testing.T) {
	for _, title := range []string{"Gaming", "Work \U0001f3ae"} {
		t.Run(title, func(t *testing.T) {
			f := newFixture(t)
			f.Exchanges = []keyboard.Exchange{{
				Write: profileTextPackets(0x70, title)[0], Response: []byte{0x54, 0xe4, 8},
			}}
			request := make([]byte, 33)
			request[0], request[1] = 0x52, 0x80
			reply := profileTextPackets(0x80, title)[0]
			reply[0] = 0x54
			f.Exchanges = append(f.Exchanges, keyboard.Exchange{Write: request, Response: reply})
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
				if err := k.SetProfile(context.Background(), title); err != nil {
					return err
				}
				got, err := k.GetProfile(context.Background())
				if got != title {
					t.Errorf("title=%q, want %q", got, title)
				}
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestApplyProfile(t *testing.T) {
	readErr := errors.New("title read failure")
	for _, tt := range []struct {
		name        string
		titleReply  []byte
		titleErr    error
		bindReply   []byte
		wantCause   error
		wantTitle   bool
		wantPartial bool
	}{
		{"success", []byte{0x54, 0xe4, 8}, nil, []byte{0x54, 0xe4, 8}, nil, true, false},
		{"title rejected", []byte{0x54, 0xe4, 9}, nil, nil, client.ErrDeviceRejected, false, false},
		{"unknown title status", []byte{0x54, 0xe4, 7}, nil, nil, client.ErrMalformedResponse, false, false},
		{"title transport", nil, readErr, nil, readErr, false, false},
		{"partial binding", []byte{0x54, 0xe4, 8}, nil, []byte{0x54, 0xe4, 9}, client.ErrDeviceRejected, true, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			profile := Profile{Title: "Gaming", Keys: []client.BindingChange{
				{Key: client.KeyA, Action: client.ActionB()},
				{Key: client.KeyC, Action: client.ActionDisabled()},
			}}
			before := append([]client.BindingChange(nil), profile.Keys...)
			f.Exchanges = []keyboard.Exchange{{
				Write: profileTextPackets(0x70, "Gaming")[0], Response: tt.titleReply, ReadErr: tt.titleErr,
			}}
			if tt.titleErr == nil && tt.titleReply[2] == 8 {
				first := bindingPacket()
				second := bindingPacket()
				second[8], second[11] = byte(client.KeyC), 0
				f.Exchanges = append(f.Exchanges,
					keyboard.Exchange{Write: first, Response: []byte{0x54, 0xe4, 8}},
					keyboard.Exchange{Write: second, Response: tt.bindReply})
			}
			err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
				return Apply(context.Background(), k, profile)
			})
			if tt.wantCause == nil {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var apply *ApplyError
				if !errors.Is(err, tt.wantCause) || !errors.As(err, &apply) || apply.TitleSet != tt.wantTitle {
					t.Fatalf("error=%v", err)
				}
				if tt.wantPartial {
					var bindings *client.ApplyError
					if !errors.As(err, &bindings) || len(bindings.Completed) != 1 || bindings.Failed != profile.Keys[1] {
						t.Fatalf("partial bindings: %v", err)
					}
				}
			}
			if !reflect.DeepEqual(before, profile.Keys) {
				t.Fatal("caller profile mutated")
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProfileValidationBeforeIO(t *testing.T) {
	for _, profile := range []Profile{
		{Title: ""}, {Title: " "}, {Title: "a\x00b"}, {Title: "\xff"},
		{Title: strings.Repeat("a", 30)}, {Title: strings.Repeat("\U0001f3ae", 15)},
		{Title: "Gaming", Keys: []client.BindingChange{{Key: 0, Action: client.ActionB()}}},
		{Title: "Gaming", Keys: []client.BindingChange{{Key: client.KeyA}}},
		{Title: "Gaming", Keys: []client.BindingChange{{Key: client.KeyA, Action: client.Action{7, 0, 0xff}}}},
		{Title: "Gaming", Keys: []client.BindingChange{
			{Key: client.KeyA, Action: client.ActionB()}, {Key: client.KeyA, Action: client.ActionC()},
		}},
	} {
		f := newFixture(t)
		err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
			return Apply(context.Background(), k, profile)
		})
		if err == nil || len(f.Writes) != 0 {
			t.Fatalf("invalid profile performed I/O: %+v, %v", profile, err)
		}
	}
}

func TestGetProfileMalformed(t *testing.T) {
	for _, reply := range [][]byte{
		nil, {0x54}, {0x54, 0x80, 2}, {0x54, 0x81, 2, 0, 0, 65},
		{0x54, 0x80, 3, 0, 0, 65, 0}, {0x54, 0x80, 2, 0, 0xd8, 0},
		{0x54, 0x80, 2, 0, 0, 0}, {0x54, 0x80, 2, 2, 0, 65},
	} {
		f := newFixture(t)
		request := make([]byte, 33)
		request[0], request[1] = 0x52, 0x80
		f.Exchanges = []keyboard.Exchange{{Write: request, Response: reply}}
		err := f.Keyboard.WithKeyboard(context.Background(), f.Device.ID, func(k client.Keyboard) error {
			_, err := k.GetProfile(context.Background())
			return err
		})
		if !errors.Is(err, client.ErrMalformedResponse) {
			t.Fatalf("reply=%x, error=%v", reply, err)
		}
		if err := f.AssertComplete(); err != nil {
			t.Fatal(err)
		}
	}
}
func TestCaptureAndApplyStoredProfile(t *testing.T) {
	f := newFixture(t)
	s := testStore(t)
	request := make([]byte, 33)
	request[0], request[1] = 0x52, 0x80
	reply := profileTextPackets(0x80, "Gaming")[0]
	reply[0] = 0x54
	mappedRequest, mappedReply := make([]byte, 33), make([]byte, 33)
	copy(mappedRequest, []byte{0x52, 0x81})
	copy(mappedReply, []byte{0x54, 0x81, 4, 7, 6, 7})
	actionRequest, actionReply := make([]byte, 33), make([]byte, 33)
	copy(actionRequest, []byte{0x52, 0x83, 4})
	copy(actionReply, []byte{0x54, 0x83, 4, 7, 0, 5})
	disabledRequest, disabledReply := make([]byte, 33), make([]byte, 33)
	copy(disabledRequest, []byte{0x52, 0x83, 6})
	copy(disabledReply, []byte{0x54, 0x83, 6, 7, 0, 0})
	f.Exchanges = []keyboard.Exchange{
		{Write: request, Response: reply},
		{Write: mappedRequest, Response: mappedReply},
		{Write: actionRequest, Response: actionReply},
		{Write: disabledRequest, Response: disabledReply},
	}
	ctx := context.Background()
	var captured Profile
	err := f.Keyboard.WithKeyboard(ctx, f.Device.ID, func(k client.Keyboard) error {
		var err error
		captured, err = Capture(ctx, k)
		return err
	})
	if err != nil || captured.Title != "Gaming" || len(captured.Keys) != 2 ||
		captured.Keys[0] != (client.BindingChange{Key: client.KeyA, Action: client.ActionB()}) ||
		captured.Keys[1] != (client.BindingChange{Key: client.KeyC, Action: client.ActionDisabled()}) {
		t.Fatalf("capture=%+v, error=%v", captured, err)
	}
	if err := s.Create(captured); err != nil {
		t.Fatal(err)
	}
	profile, err := s.Load("Gaming")
	if err != nil {
		t.Fatal(err)
	}
	disable := bindingPacket()
	disable[8], disable[11] = 6, 0
	f.Exchanges = append(f.Exchanges,
		keyboard.Exchange{Write: profileTextPackets(0x70, "Gaming")[0], Response: []byte{0x54, 0xe4, 8}},
		keyboard.Exchange{Write: bindingPacket(), Response: []byte{0x54, 0xe4, 8}},
		keyboard.Exchange{Write: disable, Response: []byte{0x54, 0xe4, 8}})
	if err := f.Keyboard.WithKeyboard(ctx, f.Device.ID, func(k client.Keyboard) error {
		return Apply(ctx, k, profile)
	}); err != nil {
		t.Fatal(err)
	}
	stored, err := s.Load("Gaming")
	if err != nil || !reflect.DeepEqual(stored, captured) {
		t.Fatalf("apply modified local profile: %+v, %v", stored, err)
	}
}

func TestEmptyProfileAndLifecycle(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	k, err := f.Keyboard.Connect(ctx, f.Device.ID)
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := k.GetProfile(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read: %v", err)
	}
	if err := Apply(cancelled, k, Profile{Title: "Gaming"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled apply: %v", err)
	}
	if err := k.SetProfile(ctx, ""); !errors.Is(err, client.ErrInvalidProfileName) {
		t.Fatalf("invalid title: %v", err)
	}
	f.Exchanges = []keyboard.Exchange{{
		Write: profileTextPackets(0x70, "Gaming")[0], Response: []byte{0x54, 0xe4, 8},
	}}
	if err := Apply(ctx, k, Profile{Title: "Gaming"}); err != nil {
		t.Fatal(err)
	}
	if err := k.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := k.GetProfile(ctx); !errors.Is(err, client.ErrDeviceClosed) {
		t.Fatalf("closed read: %v", err)
	}
	if err := k.SetProfile(ctx, "Gaming"); !errors.Is(err, client.ErrDeviceClosed) {
		t.Fatalf("closed write: %v", err)
	}
}
