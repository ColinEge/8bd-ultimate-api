package main

import (
	"fmt"
	"strings"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
)

type namedAction struct {
	name   string
	action client.Action
}

// Use the public helpers so CLI names and encodings match the Go API.
var namedActions = []namedAction{
	{"A", client.ActionA()}, {"B", client.ActionB()}, {"C", client.ActionC()},
	{"D", client.ActionD()}, {"E", client.ActionE()}, {"F", client.ActionF()},
	{"G", client.ActionG()}, {"H", client.ActionH()}, {"I", client.ActionI()},
	{"J", client.ActionJ()}, {"K", client.ActionK()}, {"L", client.ActionL()},
	{"M", client.ActionM()}, {"N", client.ActionN()}, {"O", client.ActionO()},
	{"P", client.ActionP()}, {"Q", client.ActionQ()}, {"R", client.ActionR()},
	{"S", client.ActionS()}, {"T", client.ActionT()}, {"U", client.ActionU()},
	{"V", client.ActionV()}, {"W", client.ActionW()}, {"X", client.ActionX()},
	{"Y", client.ActionY()}, {"Z", client.ActionZ()},
	{"0", client.Action0()}, {"1", client.Action1()}, {"2", client.Action2()},
	{"3", client.Action3()}, {"4", client.Action4()}, {"5", client.Action5()},
	{"6", client.Action6()}, {"7", client.Action7()}, {"8", client.Action8()},
	{"9", client.Action9()},
	{"F1", client.ActionF1()}, {"F2", client.ActionF2()}, {"F3", client.ActionF3()},
	{"F4", client.ActionF4()}, {"F5", client.ActionF5()}, {"F6", client.ActionF6()},
	{"F7", client.ActionF7()}, {"F8", client.ActionF8()}, {"F9", client.ActionF9()},
	{"F10", client.ActionF10()}, {"F11", client.ActionF11()}, {"F12", client.ActionF12()},
	{"F13", client.ActionF13()}, {"F14", client.ActionF14()}, {"F15", client.ActionF15()},
	{"F16", client.ActionF16()}, {"F17", client.ActionF17()}, {"F18", client.ActionF18()},
	{"F19", client.ActionF19()}, {"F20", client.ActionF20()}, {"F21", client.ActionF21()},
	{"F22", client.ActionF22()}, {"F23", client.ActionF23()}, {"F24", client.ActionF24()},
	{"Enter", client.ActionEnter()}, {"Esc", client.ActionEsc()},
	{"Backspace", client.ActionBackspace()}, {"Tab", client.ActionTab()},
	{"Space", client.ActionSpace()}, {"CapsLock", client.ActionCapsLock()},
	{"PrintScreen", client.ActionPrintScreen()}, {"ScrollLock", client.ActionScrollLock()},
	{"Pause", client.ActionPause()}, {"Insert", client.ActionInsert()},
	{"Home", client.ActionHome()}, {"PageUp", client.ActionPageUp()},
	{"Delete", client.ActionDelete()}, {"End", client.ActionEnd()},
	{"PageDown", client.ActionPageDown()}, {"Right", client.ActionRight()},
	{"Left", client.ActionLeft()}, {"Down", client.ActionDown()}, {"Up", client.ActionUp()},
	{"NumLock", client.ActionNumLock()}, {"NumpadDivide", client.ActionNumpadDivide()},
	{"NumpadMultiply", client.ActionNumpadMultiply()}, {"NumpadMinus", client.ActionNumpadMinus()},
	{"NumpadPlus", client.ActionNumpadPlus()}, {"NumpadEnter", client.ActionNumpadEnter()},
	{"Numpad0", client.ActionNumpad0()}, {"Numpad1", client.ActionNumpad1()},
	{"Numpad2", client.ActionNumpad2()}, {"Numpad3", client.ActionNumpad3()},
	{"Numpad4", client.ActionNumpad4()}, {"Numpad5", client.ActionNumpad5()},
	{"Numpad6", client.ActionNumpad6()}, {"Numpad7", client.ActionNumpad7()},
	{"Numpad8", client.ActionNumpad8()}, {"Numpad9", client.ActionNumpad9()},
	{"NumpadDot", client.ActionNumpadDot()},
	{"LeftCtrl", client.ActionLeftCtrl()}, {"LeftShift", client.ActionLeftShift()},
	{"LeftAlt", client.ActionLeftAlt()}, {"LeftMeta", client.ActionLeftMeta()},
	{"RightCtrl", client.ActionRightCtrl()}, {"RightShift", client.ActionRightShift()},
	{"RightAlt", client.ActionRightAlt()}, {"RightMeta", client.ActionRightMeta()},
	{"Fn", client.ActionFn()}, {"Menu", client.ActionMenu()},
	{"Grave", client.ActionGrave()}, {"Tilde", client.ActionTilde()},
	{"Exclamation", client.ActionExclamation()}, {"At", client.ActionAt()},
	{"Hash", client.ActionHash()}, {"Dollar", client.ActionDollar()},
	{"Percent", client.ActionPercent()}, {"Caret", client.ActionCaret()},
	{"Ampersand", client.ActionAmpersand()}, {"Asterisk", client.ActionAsterisk()},
	{"LeftParen", client.ActionLeftParen()}, {"RightParen", client.ActionRightParen()},
	{"Minus", client.ActionMinus()}, {"Equal", client.ActionEqual()},
	{"Underscore", client.ActionUnderscore()}, {"Plus", client.ActionPlus()},
	{"LeftBracket", client.ActionLeftBracket()}, {"RightBracket", client.ActionRightBracket()},
	{"Backslash", client.ActionBackslash()}, {"LeftBrace", client.ActionLeftBrace()},
	{"RightBrace", client.ActionRightBrace()}, {"Pipe", client.ActionPipe()},
	{"Semicolon", client.ActionSemicolon()}, {"Apostrophe", client.ActionApostrophe()},
	{"Colon", client.ActionColon()}, {"DoubleQuote", client.ActionDoubleQuote()},
	{"Comma", client.ActionComma()}, {"Dot", client.ActionDot()}, {"Slash", client.ActionSlash()},
	{"LessThan", client.ActionLessThan()}, {"GreaterThan", client.ActionGreaterThan()},
	{"Question", client.ActionQuestion()},
	{"PrimaryClick", client.ActionPrimaryClick()}, {"RightClick", client.ActionRightClick()},
	{"ScrollClick", client.ActionScrollClick()}, {"MouseForward", client.ActionMouseForward()},
	{"MouseBack", client.ActionMouseBack()}, {"DoubleClick", client.ActionDoubleClick()},
	{"ScrollUp", client.ActionScrollUp()}, {"ScrollDown", client.ActionScrollDown()},
	{"ScrollLeft", client.ActionScrollLeft()}, {"ScrollRight", client.ActionScrollRight()},
	{"VolumeUp", client.ActionVolumeUp()}, {"VolumeDown", client.ActionVolumeDown()},
	{"Mute", client.ActionMute()}, {"PlayPause", client.ActionPlayPause()},
	{"PreviousTrack", client.ActionPreviousTrack()}, {"NextTrack", client.ActionNextTrack()},
	{"Mail", client.ActionMail()}, {"FileExplorer", client.ActionFileExplorer()},
	{"Calculator", client.ActionCalculator()}, {"Browser", client.ActionBrowser()},
	{"DisplayBrightnessDown", client.ActionDisplayBrightnessDown()},
	{"DisplayBrightnessUp", client.ActionDisplayBrightnessUp()},
	{"Disabled", client.ActionDisabled()},
}

var modifiers = []struct {
	name  string
	value client.Modifier
}{
	{"LeftCtrl", client.LeftCtrl}, {"LeftShift", client.LeftShift},
	{"LeftAlt", client.LeftAlt}, {"LeftMeta", client.LeftMeta},
	{"RightCtrl", client.RightCtrl}, {"RightShift", client.RightShift},
	{"RightAlt", client.RightAlt}, {"RightMeta", client.RightMeta},
}

func findAction(name string) (client.Action, bool) {
	for _, entry := range namedActions {
		if strings.EqualFold(name, entry.name) {
			return entry.action, true
		}
	}
	return client.Action{}, false
}

func parseAction(text string) (client.Action, error) {
	if action, ok := findAction(text); ok {
		return action, nil
	}
	modifierText, actionText, chord := strings.Cut(text, "+")
	if chord {
		for _, modifier := range modifiers {
			if strings.EqualFold(modifierText, modifier.name) {
				if action, ok := findAction(actionText); ok {
					return client.Chord(modifier.value, action)
				}
				break
			}
		}
	}
	return client.Action{}, fmt.Errorf("%w: unknown action %q; use actions DEVICE (chords use LeftCtrl+C)", client.ErrInvalidAction, text)
}

func parseKey(text string) (client.KeyID, error) {
	switch strings.ToLower(text) {
	case "supera":
		return client.KeySuperA, nil
	case "superb":
		return client.KeySuperB, nil
	}
	if action, ok := findAction(text); ok {
		if key, ok := keydefs.ActionToPhysicalKey(action); ok {
			return key, nil
		}
	}
	return 0, fmt.Errorf("%w: unknown physical key %q; use keys DEVICE", client.ErrInvalidKey, text)
}

func actionName(action client.Action) string {
	switch action {
	case (client.Action{7, 0xe9, 0}):
		return "SuperA"
	case (client.Action{7, 0xe8, 0}):
		return "SuperB"
	}
	for _, entry := range namedActions {
		if entry.action == action {
			return entry.name
		}
	}
	if action.IsKeyboardUsage() && action[2] != 0 {
		for _, modifier := range modifiers {
			if byte(modifier.value) == action[1] {
				base := action
				base[1] = 0
				for _, entry := range namedActions {
					if entry.action == base {
						return modifier.name + "+" + entry.name
					}
				}
			}
		}
	}
	return fmt.Sprintf("unknown (%s)", action)
}

func keyName(key client.KeyID) string {
	return actionName(keydefs.PhysicalKeyToAction(byte(key)))
}
