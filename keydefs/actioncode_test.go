package keydefs_test

import (
	"bytes"
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/keydefs"
)

func TestParseActionCodeBytes(t *testing.T) {
	for _, tt := range []struct {
		name    string
		input   []byte
		want    keydefs.ActionCode
		invalid bool
	}{
		{name: "three bytes", input: []byte{7, 0, 5}, want: keydefs.ActionCode{7, 0, 5}},
		{name: "modifier", input: []byte{7, 0xe0, 0}, want: keydefs.ActionCode{7, 0xe0, 0}},
		{name: "combination", input: []byte{7, 0xe0, 6}, want: keydefs.ActionCode{7, 0xe0, 6}},
		{name: "six bytes", input: []byte{1, 2, 3, 4, 5, 6}, want: keydefs.ActionCode{1, 2, 3, 4, 5, 6}},
		{name: "padded", input: append([]byte{7, 0, 5}, make([]byte, 27)...), want: keydefs.ActionCode{7, 0, 5}},
		{name: "zero action", input: []byte{0, 0, 0}},
		{name: "empty", invalid: true},
		{name: "short", input: []byte{7, 0}, invalid: true},
		{name: "too long", input: []byte{1, 2, 3, 4, 5, 6, 7}, invalid: true},
		{name: "nonzero beyond padding", input: append(make([]byte, 10), 1), invalid: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			before := bytes.Clone(tt.input)
			got, err := keydefs.ParseActionCodeBytes(tt.input)
			if (err != nil) != tt.invalid || got != tt.want {
				t.Fatalf("got %v, %v; want %v, invalid=%t", got, err, tt.want, tt.invalid)
			}
			if !bytes.Equal(before, tt.input) {
				t.Fatal("parser modified its input")
			}
		})
	}
}
