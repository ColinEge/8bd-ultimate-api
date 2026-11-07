package steps

import (
	"context"
	"errors"
	"fmt"

	"github.com/ColinEge/8bd-kbd-api/client"
	"github.com/ColinEge/8bd-kbd-api/test/fixtures/keyboard"
	"github.com/cucumber/godog"
)

func registerRebindingSteps(sc *godog.ScenarioContext, w *World) {
	sc.Step(`^"([^"]*)" is the selected profile$`, w.xIsTheSelectedProfile)
	sc.Step(`^the physical key "([^"]*)" is mapped to "([^"]*)"$`, w.thePhysicalKeyXIsMappedToY)
	sc.Step(`^I have changed multiple key mappings$`, w.iHaveAValidConfiguration)
	sc.Step(`^I map the physical key "([^"]*)" to "([^"]*)"$`, w.iMapThePhysicalKeyXToY)
	sc.Step(`^I restore its default mapping$`, w.iRestoreItsDefaultMapping)
	sc.Step(`^I disable the physical key "([^"]*)"$`, w.iDisableThePhysicalKeyX)
	sc.Step(`^I attempt to map a physical key to an unsupported key$`, w.iAttemptToMapAPhysicalKeyToAnUnsupportedKey)
	sc.Step(`^I apply the selected profile$`, w.iApplyTheSelectedProfile)
	sc.Step(`^I apply the changed key mappings$`, w.iApplyTheConfiguration)
	sc.Step(`^the mapping operation should fail$`, w.theMappingOperationShouldFail)
	sc.Step(`^pressing the physical key "([^"]*)" should produce (?:"([^"]*)"|(no input))$`, w.pressingThePhysicalKeyXShouldProduceY)
	sc.Step(`^all updated mappings should be written to the keyboard$`, w.theConfigurationShouldBeWrittenToTheKeyboard)
	sc.Step(`^the physical key "([^"]*)" should read back "([^"]*)"$`, w.thePhysicalKeyXShouldReadBackY)
	sc.Step(`^the physical key "([^"]*)" should read back disabled$`, w.thePhysicalKeyXShouldReadBackDisabled)
	sc.Step(`^the physical key "A" should remain mapped to "B"$`, w.thePhysicalKeyXShouldRemainMappedToY)
	sc.Step(`^I read the keyboard bindings$`, w.iReadTheKeyboardBindings)
	sc.Step(`^the returned bindings should contain the stored mapping$`, w.theReturnedBindingsShouldContainTheStoredMapping)
}

func physicalKey(name string) (client.KeyID, error) {
	switch name {
	case "A":
		return client.KeyA, nil
	case "B":
		return client.KeyB, nil
	case "C":
		return client.KeyC, nil
	default:
		return 0, fmt.Errorf("unknown physical key %q", name)
	}
}

func namedAction(name string) (client.Action, error) {
	switch name {
	case "A":
		return client.ActionA(), nil
	case "B":
		return client.ActionB(), nil
	case "C":
		return client.ActionC(), nil
	case "Ctrl+C":
		return client.Chord(client.LeftCtrl, client.ActionC())
	case "F13":
		return client.ActionF13(), nil
	default:
		return client.Action{}, fmt.Errorf("unknown action %q", name)
	}
}

// Expected packets use independent bytes rather than the constructors under test.
func expectedNamedAction(name string) (client.Action, error) {
	switch name {
	case "A":
		return client.Action{7, 0, 4}, nil
	case "B":
		return client.Action{7, 0, 5}, nil
	case "C":
		return client.Action{7, 0, 6}, nil
	case "Ctrl+C":
		return client.Action{7, 0xe0, 6}, nil
	case "F13":
		return client.Action{7, 0, 0x68}, nil
	default:
		return client.Action{}, fmt.Errorf("unknown expected action %q", name)
	}
}

func (w *World) xIsTheSelectedProfile(string) error { return godog.ErrPending }

func (w *World) thePhysicalKeyXIsMappedToY(key, action string) error {
	if err := w.iMapThePhysicalKeyXToY(key, action); err != nil {
		return err
	}
	return w.Err
}

func (w *World) iMapThePhysicalKeyXToY(key, name string) error {
	id, err := physicalKey(key)
	if err != nil {
		return err
	}
	action, err := namedAction(name)
	if err != nil {
		return err
	}
	expected, err := expectedNamedAction(name)
	if err != nil {
		return err
	}
	w.SelectedKey = id
	w.Keyboard.Exchanges = append(w.Keyboard.Exchanges, keyboard.Exchange{
		Write: coreBindPacket(client.BindingChange{Key: id, Action: expected}), Response: []byte{0x54, 0xe4, 8},
	})
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		return k.Bind(context.Background(), id, action)
	})
	return nil
}

func (w *World) iRestoreItsDefaultMapping() {
	key := w.SelectedKey
	w.Keyboard.Exchanges = append(w.Keyboard.Exchanges, keyboard.Exchange{
		Write: coreBindPacket(client.BindingChange{Key: key, Action: client.Action{7, 0, byte(key)}}), Response: []byte{0x54, 0xe4, 8},
	})
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		return k.Restore(context.Background(), key)
	})
}

func (w *World) iDisableThePhysicalKeyX(name string) error {
	key, err := physicalKey(name)
	if err != nil {
		return err
	}
	w.SelectedKey = key
	w.Keyboard.Exchanges = append(w.Keyboard.Exchanges, keyboard.Exchange{
		Write: coreBindPacket(client.BindingChange{Key: key, Action: client.Action{7, 0, 0}}), Response: []byte{0x54, 0xe4, 8},
	})
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		return k.Disable(context.Background(), key)
	})
	return nil
}

func (w *World) iAttemptToMapAPhysicalKeyToAnUnsupportedKey() {
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		return k.Bind(context.Background(), client.KeyA, client.Action{255, 1, 2})
	})
}

func (w *World) iApplyTheSelectedProfile() error { return godog.ErrPending }

func (w *World) theMappingOperationShouldFail() error {
	if !errors.Is(w.Err, client.ErrUnsupported) {
		return fmt.Errorf("invalid mapping returned %v", w.Err)
	}
	if len(w.Keyboard.Writes) != 0 || w.Keyboard.Reads != 0 {
		return fmt.Errorf("invalid mapping performed I/O")
	}
	return nil
}

func (w *World) pressingThePhysicalKeyXShouldProduceY(string, string, string) error {
	return godog.ErrPending
}

func (w *World) thePhysicalKeyXShouldReadBackY(name, action string) error {
	key, err := physicalKey(name)
	if err != nil {
		return err
	}
	expected, err := expectedNamedAction(action)
	if err != nil {
		return err
	}
	return w.checkBinding(key, expected, client.BindingAction)
}

func (w *World) thePhysicalKeyXShouldReadBackDisabled(name string) error {
	key, err := physicalKey(name)
	if err != nil {
		return err
	}
	return w.checkBinding(key, client.Action{7, 0, 0}, client.BindingDisabled)
}

func (w *World) checkBinding(key client.KeyID, action client.Action, state client.BindingState) error {
	if w.Err != nil {
		return w.Err
	}
	if err := w.Keyboard.AssertComplete(); err != nil {
		return err
	}
	w.Keyboard.Exchanges = append(w.Keyboard.Exchanges, keyboard.Exchange{
		Write: coreActionPacket(key), Response: coreActionReply(key, action),
	})
	err := w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		got, err := k.Binding(context.Background(), key)
		if err != nil {
			return err
		}
		want := client.Binding{Key: key, State: state, Action: action, Effective: action}
		if got != want {
			return fmt.Errorf("binding=%+v; expected %+v", got, want)
		}
		return nil
	})
	return errors.Join(err, w.Keyboard.AssertComplete())
}

func (w *World) thePhysicalKeyXShouldRemainMappedToY() error {
	return w.thePhysicalKeyXShouldReadBackY("A", "B")
}

func (w *World) iReadTheKeyboardBindings() error {
	response := make([]byte, 33)
	copy(response, []byte{0x54, 0x81, 4, 7})
	request := make([]byte, 33)
	copy(request, []byte{0x52, 0x81})
	w.Keyboard.Exchanges = append(w.Keyboard.Exchanges,
		keyboard.Exchange{Write: request, Response: response},
		keyboard.Exchange{Write: coreActionPacket(client.KeyA), Response: coreActionReply(client.KeyA, client.Action{7, 0, 5})},
	)
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		var err error
		w.Bindings, err = k.Bindings(context.Background())
		return err
	})
	return w.Err
}

func (w *World) theReturnedBindingsShouldContainTheStoredMapping() error {
	want := client.Binding{Key: client.KeyA, State: client.BindingAction, Action: client.Action{7, 0, 5}, Effective: client.Action{7, 0, 5}}
	if len(w.Bindings) != 1 || w.Bindings[0] != want {
		return fmt.Errorf("stored bindings=%+v; expected %+v", w.Bindings, want)
	}
	return w.Keyboard.AssertComplete()
}
