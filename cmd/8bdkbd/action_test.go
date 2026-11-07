package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/test/fixtures/keyboard"
)

func TestRunActionPresence(t *testing.T) {
	for _, tt := range []struct {
		name   string
		action []byte
		want   string
	}{
		{name: "unbound", want: "no mapping override\n"},
		{name: "disabled", action: []byte{7, 0, 0}, want: "07-00-00\n"},
		{name: "bound", action: []byte{7, 0, 6}, want: "07-00-06\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := &keyboard.KeyboardFixture{}
			if err := f.Reset(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := f.Close(); err != nil {
					t.Error(err)
				}
			})
			request, response := make([]byte, 33), make([]byte, 33)
			copy(request, []byte{0x52, 0x83, 4})
			copy(response, []byte{0x54, 0x83, 4})
			copy(response[3:], tt.action)
			f.Exchanges = []keyboard.Exchange{{Write: request, Response: response}}
			var out, errOut bytes.Buffer
			err := run(context.Background(), []string{"action", "--device", "001-003", "--key", "0x04"},
				&out, &errOut, func(time.Duration) client.Client { return f.Keyboard })
			if err != nil || out.String() != tt.want || errOut.Len() != 0 {
				t.Fatalf("stdout=%q stderr=%q error=%v; want %q", out.String(), errOut.String(), err, tt.want)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
			if f.SessionCloses != 1 || f.ProviderCloses != 1 {
				t.Fatal("command did not close its session and provider")
			}
		})
	}
}
