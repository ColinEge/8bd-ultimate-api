package keydefs

import "testing"

func TestMain(t *testing.T) {
	type Test struct {
		actionCode ActionCode
		keyLabel   string
	}

	if err := LoadKeyMap("./testdata/bindables.yaml"); err != nil {
		t.Fatalf("Failed to load key map: %s", err.Error())
	}

	tests := []Test{}

	for actionCode, labels := range KeyMap[CategorySymbol] {
		tests = append(tests, Test{actionCode: actionCode, keyLabel: labels.KeyLabel})
	}

	for _, test := range tests {
		key, ok := KeyFromActionCode(test.actionCode)
		if !ok {
			t.Fatalf("keyFromActionCode returned not ok for %s", test.actionCode)
		}
		if key.KeyLabel != test.keyLabel {

		}
	}
}
