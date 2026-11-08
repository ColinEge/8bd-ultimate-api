package main

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
	"github.com/ColinEge/8bd-ultimate-api/test/fixtures/keyboard"
)

func TestRunRejectsInvalidArgumentsBeforeOpeningUSB(t *testing.T) {
	for _, args := range [][]string{
		{"unknown"}, {"scan", "extra"}, {"scan", "--timeout", "0"},
		{"scan", "--timeout"}, {"scan", "--timeout=bad"},
		{"binding"}, {"binding", "001-003"}, {"binding", "", "A"},
		{"binding", "001-003", "256"}, {"binding", "001-003", "F13"},
		{"binding", "001-003", "Fn"}, {"binding", "001-003", "0x04"},
		{"bind", "001-003", "A", "B"},
		{"bind", "001-003", "A", "07-00-05", "--write"},
		{"bind", "001-003", "A", "unknown", "--write"},
		{"bind", "001-003", "A", "LeftCtrl+RightAlt+C", "--write"},
		{"bind", "001-003", "A", "LeftCtrl+VolumeUp", "--write"},
		{"restore", "001-003", "A"}, {"disable", "001-003", "A"},
		{"delete-profile", "001-003"}, {"set-profile", "001-003", "Gaming"},
		{"set-profile", "001-003", "", "--write"},
		{"set-profile", "001-003", strings.Repeat("x", 30), "--write"},
		{"apply-bindings", "001-003", "A=B"},
		{"apply-bindings", "001-003", "A=B", "a=C", "--write"},
		{"apply-bindings", "001-003", "A=B", "C=unknown", "--write"},
		{"apply-bindings", "001-003", "A", "--write"},
		{"apply-bindings", "001-003", "--write"},
		{"bind", "--device", "001-003", "--key", "4", "--action", "07-00-05", "--write"},
		{"scan", "--write"}, {"info", "001-003", "extra"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out bytes.Buffer
			err := run(context.Background(), args, &out, &out, func(time.Duration) client.Client {
				t.Fatal("opened client for invalid command")
				return nil
			})
			if err == nil {
				t.Fatal("invalid command succeeded")
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	argsList := [][]string{nil, {"--help"}}
	for command := range commands {
		argsList = append(argsList, []string{command, "--help"})
	}
	for _, args := range argsList {
		var out bytes.Buffer
		err := run(context.Background(), args, &out, &out, func(time.Duration) client.Client {
			t.Fatal("help opened a client")
			return nil
		})
		if err != nil || out.Len() == 0 {
			t.Fatalf("help: output=%q error=%v", out.String(), err)
		}
	}
}

func TestRunCommands(t *testing.T) {
	for _, command := range []string{"scan", "info", "binding", "bind", "disable", "restore", "apply-bindings", "get-profile", "set-profile", "delete-profile"} {
		t.Run(command, func(t *testing.T) {
			f := &keyboard.KeyboardFixture{}
			if err := f.Reset(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := f.Close(); err != nil {
					t.Error(err)
				}
			})
			args := []string{command, "--timeout", "2s"}
			expected := "001-003\t8BitDo Retro Mechanical Keyboard\tvendor=0x2dc8 product=0x5200\n"
			if command != "scan" {
				args = append(args, "001-003")
				packet := make([]byte, 33)
				response := []byte{0x54, 0xe4, 8}
				switch command {
				case "info":
				case "binding":
					args = append(args, "A")
					copy(packet, []byte{0x52, 0x83, 4})
					response = []byte{0x54, 0x83, 4, 7, 0, 5}
					expected = "A\taction\tB\n"
				case "bind":
					args = append(args, "A", "B", "--write")
					copy(packet, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 5})
				case "disable":
					args = append(args, "A", "--write")
					copy(packet, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 0})
				case "restore":
					args = append(args, "A", "--write")
					copy(packet, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 4})
				case "apply-bindings":
					args = append(args, "A=B", "C=LeftCtrl+C", "--write")
					copy(packet, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 5})
				case "get-profile":
					copy(packet, []byte{0x52, 0x80})
					reply := make([]byte, 33)
					copy(reply, []byte{0x54, 0x80, 2, 0, 0, 'G'})
					response = reply
					expected = "G\n"
				case "set-profile":
					args = append(args, "G", "--write")
					copy(packet, []byte{0x52, 0x70, 2, 0, 0, 'G'})
				case "delete-profile":
					args = append(args, "--write")
					copy(packet, []byte{0x52, 0x70, 0, 0})
				}
				if command != "info" {
					f.Exchanges = []keyboard.Exchange{{Write: packet, Response: response}}
				}
				if command == "apply-bindings" {
					second := make([]byte, 33)
					copy(second, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 6, 7, 0xe0, 6})
					f.Exchanges = append(f.Exchanges, keyboard.Exchange{Write: second, Response: response})
				}
				if commands[command].writes {
					expected = command + " completed for device 001-003\n"
				}
			}
			var out, errOut bytes.Buffer
			err := run(context.Background(), args, &out, &errOut, func(timeout time.Duration) client.Client {
				if timeout != 2*time.Second {
					t.Errorf("timeout=%v, want 2s", timeout)
				}
				return f.Keyboard
			})
			if err != nil || out.String() != expected || errOut.Len() != 0 {
				t.Fatalf("output=%q error=%v", out.String(), err)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
			wantSessions := 1
			if command == "scan" {
				wantSessions = 0
			}
			if f.ProviderCloses != 1 || f.SessionCloses != wantSessions {
				t.Fatalf("provider closes=%d session closes=%d", f.ProviderCloses, f.SessionCloses)
			}
		})
	}
}

func TestRunPropagatesErrors(t *testing.T) {
	for _, failure := range []string{"scan", "connect", "read", "session close", "provider close"} {
		t.Run(failure, func(t *testing.T) {
			f := &keyboard.KeyboardFixture{}
			if err := f.Reset(); err != nil {
				t.Fatal(err)
			}
			cause := errors.New(failure + " failed")
			args := []string{"binding", "001-003", "A"}
			packet := make([]byte, 33)
			copy(packet, []byte{0x52, 0x83, 4})
			f.Exchanges = []keyboard.Exchange{{Write: packet, Response: []byte{0x54, 0x83, 4, 7, 0, 5}}}
			switch failure {
			case "scan":
				args = []string{"scan"}
				f.Exchanges = nil
				f.Err = cause
			case "connect":
				f.Exchanges = nil
				f.OpenErr = cause
			case "read":
				f.Exchanges[0].ReadErr = cause
			case "session close":
				f.SessionCloseErr = cause
			case "provider close":
				f.ProviderCloseErr = cause
			}
			var out bytes.Buffer
			err := run(context.Background(), args, &out, &out, func(time.Duration) client.Client { return f.Keyboard })
			if !errors.Is(err, cause) {
				t.Fatalf("got %v, want %v", err, cause)
			}
			if out.Len() != 0 {
				t.Fatalf("failed command printed success output: %q", out.String())
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
			f.ProviderCloseErr = nil
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRunCancelledBeforeOpeningClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	err := run(ctx, []string{"scan"}, &out, &out, func(time.Duration) client.Client {
		t.Fatal("cancelled command opened a client")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want cancellation", err)
	}
}

func TestRunCatalogs(t *testing.T) {
	for _, command := range []string{"keys", "actions"} {
		t.Run(command, func(t *testing.T) {
			f := newCLIFixture(t)
			var out, errOut bytes.Buffer
			err := run(context.Background(), []string{command, "001-003"}, &out, &errOut,
				func(time.Duration) client.Client { return f.Keyboard })
			if err != nil || errOut.Len() != 0 {
				t.Fatalf("stderr=%q error=%v", errOut.String(), err)
			}
			lines := strings.Split(strings.TrimSpace(out.String()), "\n")
			wantCount := 0
			for _, entry := range keydefs.Catalog() {
				if command == "keys" {
					if _, ok := keydefs.ActionToPhysicalKey(entry.Action); ok {
						wantCount++
					}
				} else if entry.Category != keydefs.CategoryKeyOnly {
					wantCount++
				}
			}
			if len(lines) != wantCount {
				t.Fatalf("got %d catalog rows, want %d", len(lines), wantCount)
			}
			for _, line := range lines {
				fields := strings.Split(line, "\t")
				if command == "keys" {
					if _, err := parseKey(fields[0]); err != nil {
						t.Fatalf("listed key cannot be used: %q: %v", fields[0], err)
					}
				} else if _, err := parseAction(fields[0]); err != nil {
					t.Fatalf("listed action cannot be used: %q: %v", fields[0], err)
				}
			}
			assertCLIFixtureComplete(t, f)
		})
	}
}

func TestRunBindings(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mapped bool
	}{{"empty", false}, {"mapped", true}} {
		t.Run(tt.name, func(t *testing.T) {
			f := newCLIFixture(t)
			request, response := make([]byte, 33), make([]byte, 33)
			copy(request, []byte{0x52, 0x81})
			copy(response, []byte{0x54, 0x81})
			want := ""
			if tt.mapped {
				response[2], response[3] = 4, 7
				want = "A\taction\tB\n"
			}
			f.Exchanges = []keyboard.Exchange{{Write: request, Response: response}}
			if tt.mapped {
				query := make([]byte, 33)
				copy(query, []byte{0x52, 0x83, 4})
				f.Exchanges = append(f.Exchanges, keyboard.Exchange{
					Write: query, Response: []byte{0x54, 0x83, 4, 7, 0, 5},
				})
			}
			var out, errOut bytes.Buffer
			err := run(context.Background(), []string{"bindings", "001-003"}, &out, &errOut,
				func(time.Duration) client.Client { return f.Keyboard })
			if err != nil || errOut.Len() != 0 || out.String() != want {
				t.Fatalf("stdout=%q stderr=%q error=%v; want %q", out.String(), errOut.String(), err, want)
			}
			assertCLIFixtureComplete(t, f)
		})
	}
}

func TestRunFirmwareVersion(t *testing.T) {
	f := newCLIFixture(t)
	var out, errOut bytes.Buffer
	err := run(context.Background(), []string{"firmware-version", "001-003"}, &out, &errOut,
		func(time.Duration) client.Client { return f.Keyboard })
	if !errors.Is(err, client.ErrUnsupported) || out.Len() != 0 {
		t.Fatalf("unsupported query: stdout=%q error=%v", out.String(), err)
	}
	assertCLIFixtureComplete(t, f)
}

func TestRunInterspersedOptions(t *testing.T) {
	for _, args := range [][]string{
		{"bind", "001-003", "A", "B", "--write", "--timeout", "2s"},
		{"bind", "--write", "--timeout=2s", "001-003", "A", "B"},
		{"bind", "001-003", "--timeout", "2s", "a", "--write", "b"},
		{"bind", "--timeout=2s", "--write", "--", "001-003", "A", "B"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			f := newCLIFixture(t)
			packet := make([]byte, 33)
			copy(packet, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 5})
			f.Exchanges = []keyboard.Exchange{{Write: packet, Response: []byte{0x54, 0xe4, 8}}}
			var out, errOut bytes.Buffer
			err := run(context.Background(), args, &out, &errOut, func(timeout time.Duration) client.Client {
				if timeout != 2*time.Second {
					t.Errorf("timeout=%s, want 2s", timeout)
				}
				return f.Keyboard
			})
			if err != nil || errOut.Len() != 0 {
				t.Fatalf("stderr=%q error=%v", errOut.String(), err)
			}
			assertCLIFixtureComplete(t, f)
		})
	}
}

func TestRunApplyBindingsPreservesPartialFailure(t *testing.T) {
	f := newCLIFixture(t)
	first, second := make([]byte, 33), make([]byte, 33)
	copy(first, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 5})
	copy(second, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 6, 7, 0, 7})
	cause := errors.New("write failed")
	f.Exchanges = []keyboard.Exchange{
		{Write: first, Response: []byte{0x54, 0xe4, 8}},
		{Write: second, WriteErr: cause},
	}
	var out, errOut bytes.Buffer
	err := run(context.Background(), []string{"apply-bindings", "001-003", "A=B", "C=D", "--write"},
		&out, &errOut, func(time.Duration) client.Client { return f.Keyboard })
	var applyErr *client.ApplyError
	if !errors.Is(err, cause) || !errors.As(err, &applyErr) || out.Len() != 0 {
		t.Fatalf("partial failure: stdout=%q error=%v", out.String(), err)
	}
	if !reflect.DeepEqual(applyErr.Completed, []client.BindingChange{{Key: client.KeyA, Action: client.ActionB()}}) ||
		applyErr.Failed != (client.BindingChange{Key: client.KeyC, Action: client.ActionD()}) {
		t.Fatalf("partial failure details: %+v", applyErr)
	}
	assertCLIFixtureComplete(t, f)
}

func TestCommandsCoverKeyboardAPI(t *testing.T) {
	methodCommands := map[string]string{
		"Info": "info", "FirmwareVersion": "firmware-version",
		"Keys": "keys", "Actions": "actions",
		"Binding": "binding", "Bindings": "bindings",
		"Bind": "bind", "Disable": "disable", "Restore": "restore",
		"ApplyBindings": "apply-bindings",
		"GetProfile":    "get-profile", "SetProfile": "set-profile", "DeleteProfile": "delete-profile",
	}
	api := reflect.TypeFor[client.Keyboard]()
	for i := 0; i < api.NumMethod(); i++ {
		method := api.Method(i).Name
		if method == "Close" {
			continue
		}
		command, ok := methodCommands[method]
		if _, exists := commands[command]; !ok || !exists {
			t.Errorf("public Keyboard method %s has no CLI command", method)
		}
	}
}

func newCLIFixture(t *testing.T) *keyboard.KeyboardFixture {
	t.Helper()
	f := &keyboard.KeyboardFixture{}
	if err := f.Reset(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})
	return f
}

func assertCLIFixtureComplete(t *testing.T, f *keyboard.KeyboardFixture) {
	t.Helper()
	if err := f.AssertComplete(); err != nil {
		t.Fatal(err)
	}
	if f.ProviderCloses != 1 || f.SessionCloses != 1 {
		t.Fatalf("provider closes=%d session closes=%d", f.ProviderCloses, f.SessionCloses)
	}
}
