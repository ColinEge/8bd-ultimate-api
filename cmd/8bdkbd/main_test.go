package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/test/fixtures/keyboard"
)

func TestRunRejectsInvalidArgumentsBeforeOpeningUSB(t *testing.T) {
	for _, args := range [][]string{
		{"unknown"}, {"scan", "extra"}, {"scan", "--timeout", "0"},
		{"action"}, {"action", "--device", "001-003"},
		{"action", "--device", "001-003", "--key", "256"},
		{"action", "--device", "001-003", "--key", "A"},
		{"bind", "--device", "001-003", "--key", "4", "--action", "07-00-05"},
		{"bind", "--device", "001-003", "--key", "4", "--action", "07", "--yes"},
		{"bind", "--device", "001-003", "--key", "4", "--action", "07-zz-05", "--yes"},
		{"restore", "--device", "001-003", "--key", "4"},
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
	for _, args := range [][]string{nil, {"--help"}, {"action", "--help"}} {
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
	for _, command := range []string{"scan", "action", "bind", "restore"} {
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
				args = append(args, "--device", "001-003", "--key", "0x04")
				packet := make([]byte, 33)
				response := []byte{0x54, 0xe4, 8}
				switch command {
				case "action":
					copy(packet, []byte{0x52, 0x83, 4})
					response = []byte{0x54, 0x83, 4, 7, 0, 5}
					expected = "07-00-05\n"
				case "bind":
					args = append(args, "--action", "07-00-05", "--yes")
					copy(packet, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 5})
					expected = "bind completed for device 001-003 key 0x04\n"
				case "restore":
					args = append(args, "--yes")
					copy(packet, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 4})
					expected = "restore completed for device 001-003 key 0x04\n"
				}
				f.Exchanges = []keyboard.Exchange{{Write: packet, Response: response}}
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
			args := []string{"action", "--device", "001-003", "--key", "4"}
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
