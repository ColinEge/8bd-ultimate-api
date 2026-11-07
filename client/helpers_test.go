package client_test

import (
	"os"
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
	"gopkg.in/yaml.v3"
)

func TestNamedActions(t *testing.T) {
	seen := make(map[client.Action]bool)
	check := func(helper func() client.Action, want client.Action) {
		t.Helper()
		got := helper()
		if got != want || !keydefs.SupportedAction(got) {
			t.Fatalf("helper returned %s; want supported %s", got, want)
		}
		seen[got] = true
		got[0] = 0xff
		if helper() != want {
			t.Fatal("helper returned shared mutable state")
		}
	}
	for _, group := range []struct {
		first   byte
		helpers []func() client.Action
	}{
		{4, []func() client.Action{
			client.ActionA, client.ActionB, client.ActionC, client.ActionD, client.ActionE, client.ActionF,
			client.ActionG, client.ActionH, client.ActionI, client.ActionJ, client.ActionK, client.ActionL,
			client.ActionM, client.ActionN, client.ActionO, client.ActionP, client.ActionQ, client.ActionR,
			client.ActionS, client.ActionT, client.ActionU, client.ActionV, client.ActionW, client.ActionX,
			client.ActionY, client.ActionZ,
		}},
		{0x1e, []func() client.Action{
			client.Action1, client.Action2, client.Action3, client.Action4, client.Action5,
			client.Action6, client.Action7, client.Action8, client.Action9, client.Action0,
		}},
		{0x28, []func() client.Action{
			client.ActionEnter, client.ActionEsc, client.ActionBackspace, client.ActionTab, client.ActionSpace,
			client.ActionMinus, client.ActionEqual, client.ActionLeftBracket, client.ActionRightBracket, client.ActionBackslash,
		}},
		{0x33, []func() client.Action{
			client.ActionSemicolon, client.ActionApostrophe, client.ActionGrave, client.ActionComma, client.ActionDot,
			client.ActionSlash, client.ActionCapsLock,
			client.ActionF1, client.ActionF2, client.ActionF3, client.ActionF4, client.ActionF5, client.ActionF6,
			client.ActionF7, client.ActionF8, client.ActionF9, client.ActionF10, client.ActionF11, client.ActionF12,
			client.ActionPrintScreen, client.ActionScrollLock, client.ActionPause, client.ActionInsert, client.ActionHome,
			client.ActionPageUp, client.ActionDelete, client.ActionEnd, client.ActionPageDown,
			client.ActionRight, client.ActionLeft, client.ActionDown, client.ActionUp,
			client.ActionNumLock, client.ActionNumpadDivide, client.ActionNumpadMultiply, client.ActionNumpadMinus,
			client.ActionNumpadPlus, client.ActionNumpadEnter,
			client.ActionNumpad1, client.ActionNumpad2, client.ActionNumpad3, client.ActionNumpad4, client.ActionNumpad5,
			client.ActionNumpad6, client.ActionNumpad7, client.ActionNumpad8, client.ActionNumpad9,
			client.ActionNumpad0, client.ActionNumpadDot,
		}},
		{0x68, []func() client.Action{
			client.ActionF13, client.ActionF14, client.ActionF15, client.ActionF16, client.ActionF17, client.ActionF18,
			client.ActionF19, client.ActionF20, client.ActionF21, client.ActionF22, client.ActionF23, client.ActionF24,
		}},
	} {
		for i, helper := range group.helpers {
			check(helper, client.Action{7, 0, group.first + byte(i)})
		}
	}
	for i, helper := range []func() client.Action{
		client.ActionLeftCtrl, client.ActionLeftShift, client.ActionLeftAlt, client.ActionLeftMeta,
		client.ActionRightCtrl, client.ActionRightShift, client.ActionRightAlt, client.ActionRightMeta,
	} {
		check(helper, client.Action{7, 0xe0 + byte(i), 0})
		for _, base := range []client.Action{client.ActionA(), client.ActionF24(), client.ActionMinus()} {
			got, err := client.Chord(client.Modifier(0xe0+byte(i)), base)
			want := client.Action{7, 0xe0 + byte(i), base[2]}
			if err != nil || got != want || !keydefs.SupportedAction(got) {
				t.Fatalf("chord: got %s, %v; want %s", got, err, want)
			}
		}
	}
	for _, test := range []struct {
		helper func() client.Action
		want   client.Action
	}{
		{client.ActionFn, client.Action{7, 0, 3}},
		{client.ActionMenu, client.Action{7, 0xe1, 0x43}},
		{client.ActionTilde, client.Action{7, 0xe1, 0x35}},
		{client.ActionExclamation, client.Action{7, 0xe1, 0x1e}},
		{client.ActionAt, client.Action{7, 0xe1, 0x1f}},
		{client.ActionHash, client.Action{7, 0xe1, 0x20}},
		{client.ActionDollar, client.Action{7, 0xe1, 0x21}},
		{client.ActionPercent, client.Action{7, 0xe1, 0x22}},
		{client.ActionCaret, client.Action{7, 0xe1, 0x23}},
		{client.ActionAmpersand, client.Action{7, 0xe1, 0x24}},
		{client.ActionAsterisk, client.Action{7, 0xe1, 0x25}},
		{client.ActionLeftParen, client.Action{7, 0xe1, 0x26}},
		{client.ActionRightParen, client.Action{7, 0xe1, 0x27}},
		{client.ActionUnderscore, client.Action{7, 0xe1, 0x2d}},
		{client.ActionPlus, client.Action{7, 0xe1, 0x2e}},
		{client.ActionLeftBrace, client.Action{7, 0xe1, 0x2f}},
		{client.ActionRightBrace, client.Action{7, 0xe1, 0x30}},
		{client.ActionPipe, client.Action{7, 0xe1, 0x31}},
		{client.ActionColon, client.Action{7, 0xe1, 0x33}},
		{client.ActionDoubleQuote, client.Action{7, 0xe1, 0x34}},
		{client.ActionLessThan, client.Action{7, 0xe1, 0x36}},
		{client.ActionGreaterThan, client.Action{7, 0xe1, 0x37}},
		{client.ActionQuestion, client.Action{7, 0xe1, 0x38}},
		{client.ActionPrimaryClick, client.Action{1, 1, 0}},
		{client.ActionRightClick, client.Action{1, 2, 0}},
		{client.ActionScrollClick, client.Action{1, 4, 0}},
		{client.ActionMouseForward, client.Action{1, 0x10, 0}},
		{client.ActionMouseBack, client.Action{1, 8, 0}},
		{client.ActionDoubleClick, client.Action{1, 0x80, 0}},
		{client.ActionScrollUp, client.Action{1, 0, 0, 0, 2}},
		{client.ActionScrollDown, client.Action{1, 0, 0, 0, 0xfe}},
		{client.ActionScrollLeft, client.Action{1, 0, 0, 0, 0, 0xfe}},
		{client.ActionScrollRight, client.Action{1, 0, 0, 0, 0, 2}},
		{client.ActionVolumeDown, client.Action{0x0c, 0xea, 0}},
		{client.ActionVolumeUp, client.Action{0x0c, 0xe9, 0}},
		{client.ActionMute, client.Action{0x0c, 0xe2, 0}},
		{client.ActionPlayPause, client.Action{0x0c, 0xcd, 0}},
		{client.ActionPreviousTrack, client.Action{0x0c, 0xb6, 0}},
		{client.ActionNextTrack, client.Action{0x0c, 0xb5, 0}},
		{client.ActionMail, client.Action{0x0c, 0x8a, 1}},
		{client.ActionFileExplorer, client.Action{7, 0xe3, 8}},
		{client.ActionCalculator, client.Action{0x0c, 0x92, 1}},
		{client.ActionBrowser, client.Action{0x0c, 0x23, 2}},
		{client.ActionDisplayBrightnessDown, client.Action{0x0c, 0x70, 0}},
		{client.ActionDisplayBrightnessUp, client.Action{0x0c, 0x6f, 0}},
		{client.ActionDisabled, client.Action{7, 0, 0}},
	} {
		check(test.helper, test.want)
	}
	for _, entry := range keydefs.Catalog() {
		if entry.Category != keydefs.CategoryKeyOnly && !seen[entry.Action] {
			t.Errorf("catalog action %s (%s) has no tested named helper", entry.Action, entry.Label)
		}
	}
}

func TestNamedKeysMatchDefinitions(t *testing.T) {
	data, err := os.ReadFile("../keys.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var definitions struct {
		Keys []struct {
			ID           string
			Code         byte
			DefaultUsage string `yaml:"default_usage"`
		}
		Usages []struct {
			ID   string
			Code uint32
		}
	}
	if err := yaml.Unmarshal(data, &definitions); err != nil {
		t.Fatal(err)
	}
	named := map[string]client.KeyID{
		"a": client.KeyA, "b": client.KeyB, "c": client.KeyC, "d": client.KeyD, "e": client.KeyE, "f": client.KeyF,
		"g": client.KeyG, "h": client.KeyH, "i": client.KeyI, "j": client.KeyJ, "k": client.KeyK, "l": client.KeyL,
		"m": client.KeyM, "n": client.KeyN, "o": client.KeyO, "p": client.KeyP, "q": client.KeyQ, "r": client.KeyR,
		"s": client.KeyS, "t": client.KeyT, "u": client.KeyU, "v": client.KeyV, "w": client.KeyW, "x": client.KeyX,
		"y": client.KeyY, "z": client.KeyZ,
		"1": client.Key1, "2": client.Key2, "3": client.Key3, "4": client.Key4, "5": client.Key5,
		"6": client.Key6, "7": client.Key7, "8": client.Key8, "9": client.Key9, "0": client.Key0,
		"f1": client.KeyF1, "f2": client.KeyF2, "f3": client.KeyF3, "f4": client.KeyF4, "f5": client.KeyF5,
		"f6": client.KeyF6, "f7": client.KeyF7, "f8": client.KeyF8, "f9": client.KeyF9, "f10": client.KeyF10,
		"f11": client.KeyF11, "f12": client.KeyF12,
		"esc": client.KeyEsc, "enter": client.KeyEnter, "backspace": client.KeyBackspace, "tab": client.KeyTab,
		"space": client.KeySpace, "minus": client.KeyMinus, "equal": client.KeyEqual,
		"leftbrace": client.KeyLeftBracket, "rightbrace": client.KeyRightBracket, "backslash": client.KeyBackslash,
		"semicolon": client.KeySemicolon, "apostrophe": client.KeyApostrophe, "grave": client.KeyGrave,
		"comma": client.KeyComma, "dot": client.KeyDot, "slash": client.KeySlash, "capslock": client.KeyCapsLock,
		"prtsc": client.KeyPrintScreen, "scrlk": client.KeyScrollLock, "pause": client.KeyPause,
		"insert": client.KeyInsert, "home": client.KeyHome, "pageup": client.KeyPageUp, "delete": client.KeyDelete,
		"end": client.KeyEnd, "pagedown": client.KeyPageDown,
		"right": client.KeyRight, "left": client.KeyLeft, "down": client.KeyDown, "up": client.KeyUp,
		"numlock": client.KeyNumLock, "kpslash": client.KeyNumpadDivide, "kpasterisk": client.KeyNumpadMultiply,
		"kpminus": client.KeyNumpadMinus, "kpplus": client.KeyNumpadPlus, "kpenter": client.KeyNumpadEnter,
		"kp1": client.KeyNumpad1, "kp2": client.KeyNumpad2, "kp3": client.KeyNumpad3, "kp4": client.KeyNumpad4,
		"kp5": client.KeyNumpad5, "kp6": client.KeyNumpad6, "kp7": client.KeyNumpad7, "kp8": client.KeyNumpad8,
		"kp9": client.KeyNumpad9, "kp0": client.KeyNumpad0, "kpdot": client.KeyNumpadDot,
		"leftctrl": client.KeyLeftCtrl, "leftshift": client.KeyLeftShift, "leftalt": client.KeyLeftAlt,
		"leftmeta": client.KeyLeftMeta, "rightctrl": client.KeyRightCtrl, "rightshift": client.KeyRightShift,
		"rightalt": client.KeyRightAlt,
	}
	usages := make(map[string]client.Action)
	for _, usage := range definitions.Usages {
		usages[usage.ID] = client.Action{byte(usage.Code >> 16), byte(usage.Code >> 8), byte(usage.Code)}
	}
	for _, key := range definitions.Keys {
		switch key.ID {
		case "supera", "superb", "compose", "menu", "rightmeta", "windows":
			continue
		}
		id, ok := named[key.ID]
		if !ok || byte(id) != key.Code {
			t.Errorf("key %s: got %s (present=%t); want 0x%02x", key.ID, id, ok, key.Code)
		}
		if got := keydefs.PhysicalKeyToAction(key.Code); got != usages[key.DefaultUsage] {
			t.Errorf("default for %s: got %s; want %s", key.ID, got, usages[key.DefaultUsage])
		}
		delete(named, key.ID)
	}
	if len(named) != 0 {
		t.Fatalf("named keys missing from keys.yaml: %v", named)
	}
}
