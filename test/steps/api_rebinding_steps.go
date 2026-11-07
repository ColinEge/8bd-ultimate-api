package steps

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ColinEge/8bd-kbd-api/client"
	"github.com/ColinEge/8bd-kbd-api/keydefs"
	"github.com/ColinEge/8bd-kbd-api/test/fixtures/keyboard"
	"github.com/cucumber/godog"
)

func InitializeAPIBindingScenario(sc *godog.ScenarioContext) {
	w := &World{Keyboard: &keyboard.KeyboardFixture{}}
	registerAPIBindingSteps(sc, w)
}

func registerAPIBindingSteps(sc *godog.ScenarioContext, w *World) {
	sc.Step(`^a scripted keyboard for the action "([^"]*)"$`, w.scriptedKeyboardForAction)
	sc.Step(`^I bind the selected action through the public API$`, w.bindSelectedAction)
	sc.Step(`^a scripted keyboard with a binding response of "([^"]*)"$`, w.scriptedBindingKeyboard)
	sc.Step(`^I bind A to B through the public API$`, w.bindAtoB)
	sc.Step(`^the binding result should be "([^"]*)"$`, w.bindingResult)
	sc.Step(`^the binding session should be closed$`, w.bindingSessionClosed)
	sc.Step(`^a scripted keyboard for a binding round-trip$`, w.scriptedBindingRoundTrip)
	sc.Step(`^I bind, read, and restore A through the public API$`, w.bindingRoundTrip)
	sc.Step(`^a scripted action response of "([^"]*)"$`, w.scriptedActionResponse)
	sc.Step(`^I read the scripted action through the public API$`, w.readScriptedAction)
	sc.Step(`^the action presence should be "([^"]*)"$`, w.actionPresence)
	sc.Step(`^a scripted default for modifier key "([^"]*)" at "([^"]*)" with action "([^"]*)"$`, w.scriptedModifierDefault)
	sc.Step(`^I read and restore the modifier through the public API$`, w.readAndRestoreModifier)
	sc.After(w.closeBindingFixture)
}

func (w *World) scriptedKeyboardForAction(name string) error {
	var expected client.Action
	w.Err = nil
	if err := w.Keyboard.Reset(); err != nil {
		return err
	}
	switch name {
	case "disabled":
		w.SelectedAction = client.ActionDisabled()
		expected = client.Action{7, 0, 0}
	case "Ctrl+C":
		action, err := client.Chord(client.LeftCtrl, client.ActionC())
		if err != nil {
			return err
		}
		w.SelectedAction = action
		expected = client.Action{7, 0xe0, 6}
	case "F13":
		w.SelectedAction = client.ActionF13()
		expected = client.Action{7, 0, 0x68}
	case "Z":
		w.SelectedAction = client.ActionZ()
		expected = client.Action{7, 0, 0x1d}
	case "0":
		w.SelectedAction = client.Action0()
		expected = client.Action{7, 0, 0x27}
	case "F24":
		w.SelectedAction = client.ActionF24()
		expected = client.Action{7, 0, 0x73}
	case "numpad enter":
		w.SelectedAction = client.ActionNumpadEnter()
		expected = client.Action{7, 0, 0x58}
	case "left ctrl":
		w.SelectedAction = client.ActionLeftCtrl()
		expected = client.Action{7, 0xe0, 0}
	case "right meta":
		w.SelectedAction = client.ActionRightMeta()
		expected = client.Action{7, 0xe7, 0}
	case "minus":
		w.SelectedAction = client.ActionMinus()
		expected = client.Action{7, 0, 0x2d}
	case "underscore":
		w.SelectedAction = client.ActionUnderscore()
		expected = client.Action{7, 0xe1, 0x2d}
	case "double quote":
		w.SelectedAction = client.ActionDoubleQuote()
		expected = client.Action{7, 0xe1, 0x34}
	case "scroll left":
		w.SelectedAction = client.ActionScrollLeft()
		expected = client.Action{1, 0, 0, 0, 0, 0xfe}
	case "next track":
		w.SelectedAction = client.ActionNextTrack()
		expected = client.Action{0x0c, 0xb5, 0}
	case "volume up":
		w.SelectedAction = client.ActionVolumeUp()
		expected = client.Action{0x0c, 0xe9, 0}
	case "volume down":
		w.SelectedAction = client.ActionVolumeDown()
		expected = client.Action{0x0c, 0xea, 0}
	case "mute":
		w.SelectedAction = client.ActionMute()
		expected = client.Action{0x0c, 0xe2, 0}
	default:
		return fmt.Errorf("unsupported action name %q", name)
	}
	request := make([]byte, 33)
	copy(request, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4})
	copy(request[9:], expected[:])
	w.Keyboard.Exchanges = []keyboard.Exchange{{Write: request, Response: []byte{0x54, 0xe4, 8}}}
	return nil
}

func (w *World) scriptedModifierDefault(name, keyText, actionText string) error {
	if err := w.Keyboard.Reset(); err != nil {
		return err
	}
	w.Err = nil
	switch name {
	case "left ctrl":
		w.SelectedKey = client.KeyLeftCtrl
	case "left shift":
		w.SelectedKey = client.KeyLeftShift
	case "left alt":
		w.SelectedKey = client.KeyLeftAlt
	case "left meta":
		w.SelectedKey = client.KeyLeftMeta
	case "right ctrl":
		w.SelectedKey = client.KeyRightCtrl
	case "right shift":
		w.SelectedKey = client.KeyRightShift
	case "right alt":
		w.SelectedKey = client.KeyRightAlt
	default:
		return fmt.Errorf("unsupported physical modifier %q", name)
	}
	value, err := strconv.ParseUint(keyText, 0, 8)
	if err != nil {
		return fmt.Errorf("invalid scripted modifier key %q: %w", keyText, err)
	}
	key := keydefs.KeyID(value)
	if key != w.SelectedKey {
		return fmt.Errorf("named modifier key %s, want %s", w.SelectedKey, key)
	}
	action, err := client.ParseAction(actionText)
	if err != nil {
		return err
	}
	w.ExpectedAction = action
	restore := make([]byte, 33)
	copy(restore, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, byte(key)})
	copy(restore[9:], action[:])
	w.Keyboard.Exchanges = []keyboard.Exchange{
		{Write: coreActionPacket(key), Response: coreActionReply(key, client.Action{})},
		{Write: restore, Response: []byte{0x54, 0xe4, 8}},
		{Write: coreActionPacket(key), Response: coreActionReply(key, action)},
	}
	return nil
}

func (w *World) readAndRestoreModifier() error {
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		ctx := context.Background()
		binding, err := k.Binding(ctx, w.SelectedKey)
		if err != nil {
			return err
		}
		if binding.State != client.BindingDefault || binding.Action != (client.Action{}) || binding.Effective != w.ExpectedAction {
			return fmt.Errorf("default modifier binding %+v, want effective %s", binding, w.ExpectedAction)
		}
		if err := k.Restore(ctx, w.SelectedKey); err != nil {
			return err
		}
		binding, err = k.Binding(ctx, w.SelectedKey)
		if err != nil {
			return err
		}
		if binding.State != client.BindingAction || binding.Action != w.ExpectedAction || binding.Effective != w.ExpectedAction {
			return fmt.Errorf("restored modifier binding %+v, want %s", binding, w.ExpectedAction)
		}
		return nil
	})
	return nil
}

func (w *World) bindSelectedAction() error {
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		if w.SelectedAction == client.ActionDisabled() {
			return k.Disable(context.Background(), client.KeyA)
		}
		return k.Bind(context.Background(), client.KeyA, w.SelectedAction)
	})
	return nil
}

func (w *World) scriptedActionResponse(kind string) error {
	w.Err = nil
	if err := w.Keyboard.Reset(); err != nil {
		return err
	}
	switch kind {
	case "unbound":
		w.ExpectedAction = keydefs.ActionCode{}
	case "disabled":
		w.ExpectedAction = keydefs.ActionCode{7, 0, 0}
	case "bound":
		w.ExpectedAction = keydefs.ActionCode{7, 0, 6}
	default:
		return fmt.Errorf("unsupported action response %q", kind)
	}
	request, response := make([]byte, 33), make([]byte, 33)
	copy(request, []byte{0x52, 0x83, 4})
	copy(response, []byte{0x54, 0x83, 4})
	copy(response[3:], w.ExpectedAction[:])
	w.Keyboard.Exchanges = []keyboard.Exchange{{Write: request, Response: response}}
	return nil
}

func (w *World) readScriptedAction() error {
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(s client.Keyboard) error {
		var err error
		binding, readErr := s.Binding(context.Background(), keydefs.KeyID(4))
		w.ReadAction, w.ActionMapped, err = binding.Action, readErr == nil && binding.State != client.BindingDefault, readErr
		return err
	})
	return w.Err
}

func (w *World) actionPresence(presence string) error {
	if presence != "mapped" && presence != "absent" {
		return fmt.Errorf("unsupported presence %q", presence)
	}
	if w.ReadAction != w.ExpectedAction || w.ActionMapped != (presence == "mapped") {
		return fmt.Errorf("got action=%s mapped=%t; want action=%s presence=%s", w.ReadAction, w.ActionMapped, w.ExpectedAction, presence)
	}
	return nil
}

func (w *World) closeBindingFixture(ctx context.Context, scenario *godog.Scenario, scenarioErr error) (context.Context, error) {
	return ctx, errors.Join(w.Keyboard.AssertComplete(), w.Keyboard.Close())
}

func (w *World) scriptedBindingKeyboard(response string) error {
	w.Err = nil
	if err := w.Keyboard.Reset(); err != nil {
		return err
	}
	packet := make([]byte, 33)
	copy(packet, []byte{0x52, 0xfa, 0x03, 0x0c, 0x00, 0xaa, 0x09, 0x71, 0x04, 0x07, 0x00, 0x05})
	exchange := keyboard.Exchange{Write: packet}
	switch response {
	case "ok":
		exchange.Response = []byte{0x54, 0xe4, 0x08}
	case "bad":
		exchange.Response = []byte{0x54, 0xe4, 0x09}
	case "malformed":
		exchange.Response = []byte{0x54}
	case "write failure":
		exchange.WriteErr = errors.New("scripted write failure")
	case "read failure":
		exchange.ReadErr = errors.New("scripted read failure")
	default:
		return fmt.Errorf("unsupported scripted response %q", response)
	}
	w.Keyboard.Exchanges = []keyboard.Exchange{exchange}
	return nil
}

func (w *World) bindAtoB() error {
	if err := w.Keyboard.RequireConnected(); err != nil {
		return err
	}
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(s client.Keyboard) error {
		return s.Bind(context.Background(), keydefs.KeyID(0x04), keydefs.ActionCode{0x07, 0x00, 0x05})
	})
	return nil
}

func (w *World) bindingResult(expected string) error {
	if expected == "success" {
		return w.Err
	}
	if w.Err == nil || !strings.Contains(w.Err.Error(), expected) {
		return fmt.Errorf("got %v, want binding error containing %q", w.Err, expected)
	}
	return nil
}

func (w *World) bindingSessionClosed() error {
	if w.Keyboard.SessionCloses != 1 {
		return fmt.Errorf("got %d session closes, want 1", w.Keyboard.SessionCloses)
	}
	return w.Keyboard.AssertComplete()
}

func (w *World) scriptedBindingRoundTrip() error {
	if err := w.scriptedBindingKeyboard("ok"); err != nil {
		return err
	}
	read := make([]byte, 33)
	copy(read, []byte{0x52, 0x83, 0x04})
	restore := make([]byte, 33)
	copy(restore, []byte{0x52, 0xfa, 0x03, 0x0c, 0x00, 0xaa, 0x09, 0x71, 0x04, 0x07, 0x00, 0x04})
	w.Keyboard.Exchanges = append(w.Keyboard.Exchanges,
		keyboard.Exchange{Write: read, Response: []byte{0x54, 0x83, 0x04, 0x07, 0x00, 0x05}},
		keyboard.Exchange{Write: restore, Response: []byte{0x54, 0xe4, 0x08}},
		keyboard.Exchange{Write: read, Response: []byte{0x54, 0x83, 0x04, 0x07, 0x00, 0x04}},
	)
	return nil
}

func (w *World) bindingRoundTrip() error {
	if err := w.Keyboard.RequireConnected(); err != nil {
		return err
	}
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(s client.Keyboard) error {
		ctx := context.Background()
		key := keydefs.KeyID(4)
		b := keydefs.ActionCode{7, 0, 5}
		if err := s.Bind(ctx, key, b); err != nil {
			return err
		}
		binding, err := s.Binding(ctx, key)
		action, mapped := binding.Action, binding.State != client.BindingDefault
		if err != nil {
			return err
		}
		if !mapped || action != b {
			return fmt.Errorf("binding read-back: got %s, want %s", action, b)
		}
		if err := s.Restore(ctx, key); err != nil {
			return err
		}
		binding, err = s.Binding(ctx, key)
		action, mapped = binding.Action, binding.State != client.BindingDefault
		if err != nil {
			return err
		}
		a := keydefs.ActionCode{7, 0, 4}
		if !mapped || action != a {
			return fmt.Errorf("restored read-back: got %s, want %s", action, a)
		}
		return nil
	})
	return nil
}
