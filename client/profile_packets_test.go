package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ColinEge/8bd-ultimate-api/transport"
)

func TestProfileTwoPackets(t *testing.T) {
	title := strings.Repeat("a", 14) + "\U0001f3ae" + strings.Repeat("b", 13)
	data := make([]byte, 0, 58)
	for i := 0; i < 14; i++ {
		data = append(data, 0, 'a')
	}
	data = append(data, 0xd8, 0x3c, 0xdf, 0xae)
	for i := 0; i < 13; i++ {
		data = append(data, 0, 'b')
	}
	first, second, request := make([]byte, 33), make([]byte, 33), make([]byte, 33)
	copy(first, []byte{0x52, 0x70, 29, 1})
	copy(second, []byte{0x52, 0x70, 29, 0})
	copy(first[4:], data[:29])
	copy(second[4:], data[29:])
	copy(request, []byte{0x52, 0x80})
	wantWrites := [][]byte{first, second, request}
	readFirst, readSecond := bytes.Clone(first), bytes.Clone(second)
	readFirst[0], readFirst[1] = 0x54, 0x80
	readSecond[0], readSecond[1] = 0x54, 0x80
	replies := [][]byte{{0x54, 0xe4, 8}, readFirst, readSecond}
	writes, reads := 0, 0
	k := &keyboard{
		timeout: time.Second, gate: make(chan struct{}, 1),
		transport: &transport.Mock{
			WriteFunc: func(_ context.Context, data []byte) error {
				if writes >= len(wantWrites) || !bytes.Equal(data, wantWrites[writes]) {
					return fmt.Errorf("unexpected write %x at %d", data, writes)
				}
				writes++
				return nil
			},
			ReadFunc: func(context.Context) ([]byte, error) {
				if reads >= len(replies) || (reads == 0 && writes != 2) || (reads > 0 && writes != 3) {
					return nil, fmt.Errorf("unexpected read %d after %d writes", reads, writes)
				}
				reply := replies[reads]
				reads++
				return reply, nil
			},
		},
	}
	k.gate <- struct{}{}
	ctx := context.Background()
	if err := k.SetProfile(ctx, title); err != nil {
		t.Fatal(err)
	}
	got, err := k.GetProfile(ctx)
	if err != nil || got != title {
		t.Fatalf("title=%q, error=%v", got, err)
	}
	if writes != 3 || reads != 3 {
		t.Fatalf("incomplete I/O: writes=%d, reads=%d", writes, reads)
	}
	if err := k.Close(); err != nil {
		t.Fatal(err)
	}
}
