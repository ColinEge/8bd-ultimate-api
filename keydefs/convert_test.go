package keydefs

import "testing"

func TestPhysicalKeyConversions(t *testing.T) {
	for key := byte(4); key <= 0x6b; key++ {
		want := ActionCode{7, 0, key}
		if key >= 0x64 {
			want = ActionCode{7, key + 0x7c, 0}
		}
		got := PhysicalKeyToAction(key)
		physical, ok := ActionToPhysicalKey(want)
		if got != want || !ok || byte(physical) != key {
			t.Errorf("key 0x%02x: action=%s, inverse=%s/%t; want %s", key, got, physical, ok, want)
		}
	}
	for _, action := range []ActionCode{
		{}, {7, 0, 0}, {7, 0, 1}, {7, 0, 3}, {7, 0, 0x64}, {7, 0, 0x65},
		{7, 0, 0x68}, {7, 0, 0x73}, {7, 0xdf, 0}, {7, 0xe8, 0}, {7, 0xe9, 0},
		{7, 0xff, 0}, {7, 1, 4}, {7, 0, 4, 1}, {0x0c, 0xe9, 0}, {1, 1, 0},
	} {
		if key, ok := ActionToPhysicalKey(action); ok || key != 0 {
			t.Errorf("unexpected physical conversion for %s: %s/%t", action, key, ok)
		}
	}
	for modifier := byte(0xe0); modifier <= 0xe7; modifier++ {
		if key, ok := ActionToPhysicalKey(ActionCode{7, modifier, 4}); ok || key != 0 {
			t.Errorf("chord mapped to physical key: %s/%t", key, ok)
		}
	}
	for _, key := range []byte{0, 1, 3, 0x6c, 0x6d, 0xe8, 0xe9, 0xff} {
		if got := PhysicalKeyToAction(key); got != (ActionCode{}) {
			t.Errorf("unexpected action for invalid key 0x%02x: %s", key, got)
		}
	}
}
