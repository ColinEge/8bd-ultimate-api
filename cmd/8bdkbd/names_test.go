package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
)

func TestNamedActionsCoverCatalog(t *testing.T) {
	seenNames := make(map[string]bool)
	seenActions := make(map[client.Action]bool)
	for _, entry := range namedActions {
		name := strings.ToLower(entry.name)
		if seenNames[name] || seenActions[entry.action] {
			t.Fatalf("duplicate name or action: %s", entry.name)
		}
		seenNames[name], seenActions[entry.action] = true, true
		got, err := parseAction(name)
		if err != nil || got != entry.action || actionName(got) != entry.name {
			t.Fatalf("action %s: got %s, %v", entry.name, got, err)
		}
		if !keydefs.SupportedAction(got) {
			t.Fatalf("unsupported named action: %s", entry.name)
		}
	}
	for _, entry := range keydefs.Catalog() {
		if entry.Category != keydefs.CategoryKeyOnly && !seenActions[entry.Action] {
			t.Errorf("missing named action for %s (%s)", entry.Action, entry.Label)
		}
	}
}

func TestParseKeyAndActionAreDistinct(t *testing.T) {
	for _, tt := range []struct {
		name string
		want client.KeyID
	}{
		{"A", client.KeyA}, {"4", client.Key4}, {"Enter", client.KeyEnter},
		{"F12", client.KeyF12}, {"Numpad1", client.KeyNumpad1},
		{"leftctrl", client.KeyLeftCtrl}, {"RightCtrl", client.KeyRightCtrl},
		{"RightAlt", client.KeyRightAlt}, {"Minus", client.KeyMinus},
	} {
		got, err := parseKey(tt.name)
		if err != nil || got != tt.want {
			t.Fatalf("key %s: got %s, %v; want %s", tt.name, got, err, tt.want)
		}
	}
	for _, name := range []string{"F13", "VolumeUp", "Disabled", "Fn", "LeftCtrl+C", "0x04", "256", "SuperA"} {
		if _, err := parseKey(name); !errors.Is(err, client.ErrInvalidKey) {
			t.Errorf("key %q: expected invalid key, got %v", name, err)
		}
	}
}

func TestChordNames(t *testing.T) {
	for _, modifier := range modifiers {
		for _, base := range []string{"C", "F24", "Minus"} {
			name := modifier.name + "+" + base
			action, err := parseAction(strings.ToLower(name))
			if err != nil || action[1] != byte(modifier.value) {
				t.Fatalf("chord %s: got %s, %v", name, action, err)
			}
			roundTrip, err := parseAction(actionName(action))
			if err != nil || roundTrip != action {
				t.Fatalf("chord %s does not round-trip: %s, %v", name, roundTrip, err)
			}
		}
	}
	for _, name := range []string{
		"Ctrl+C", "LeftCtrl+", "LeftCtrl+RightShift+C",
		"LeftCtrl+VolumeUp", "LeftCtrl+LeftShift", "LeftCtrl+Disabled",
		"LeftCtrl+Plus", "LeftCtrl+unknown",
	} {
		if _, err := parseAction(name); err == nil {
			t.Errorf("invalid chord %q accepted", name)
		}
	}
}
