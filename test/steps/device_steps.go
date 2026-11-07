package steps

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/test/fixtures/keyboard"
	"github.com/cucumber/godog"
)

func registerDeviceSteps(sc *godog.ScenarioContext, w *World) {
	sc.Step(`^a connected 8BitDo keyboard$`, w.an8BitDoKeyboardIsConnected)
	sc.Step(`^an 8BitDo keyboard was previously connected$`, w.an8BitDoKeyboardIsConnected)
	sc.Step(`^no keyboard is connected$`, w.noKeyboardIsConnected)
	sc.Step(`^I have a valid configuration$`, w.iHaveAValidConfiguration)
	sc.Step(`^the keyboard has been disconnected$`, w.theKeyboardHasBeenDisconnected)
	sc.Step(`^I request the connected devices$`, w.iRequestTheConnectedDevices)
	sc.Step(`^I request its device information$`, w.iRequestItsDeviceInformation)
	sc.Step(`^I request its firmware version$`, w.iRequestItsFirmwareVersion)
	sc.Step(`^the keyboard is disconnected$`, w.theKeyboardHasBeenDisconnected)
	sc.Step(`^the keyboard is reconnected$`, w.theKeyboardIsReconnected)
	sc.Step(`^I apply the configuration$`, w.iApplyTheConfiguration)
	sc.Step(`^I attempt to apply a configuration$`, w.iApplyTheConfiguration)
	sc.Step(`^the keyboard should appear in the device list$`, w.theKeyboardShouldAppearInTheDeviceList)
	sc.Step(`^its model should be reported$`, w.itsModelShouldBeReported)
	sc.Step(`^its firmware version should be reported$`, w.itsFirmwareVersionShouldBeReported)
	sc.Step(`^firmware querying should report unsupported$`, w.firmwareQueryingShouldReportUnsupported)
	sc.Step(`^the device should no longer be available$`, w.theDeviceShouldNoLongerBeAvailable)
	sc.Step(`^the device should become available again$`, w.theDeviceShouldBecomeAvailableAgain)
	sc.Step(`^the configuration should be written to the keyboard$`, w.theConfigurationShouldBeWrittenToTheKeyboard)
	sc.Step(`^the missing device error should be returned without writing$`, w.theMissingDeviceErrorShouldBeReturnedWithoutWriting)
}

func (w *World) noKeyboardIsConnected() error {
	if err := w.Keyboard.Reset(); err != nil {
		return err
	}
	w.Keyboard.Connected = false
	return nil
}

func (w *World) iHaveAValidConfiguration() {
	w.prepareChanges([]client.BindingChange{
		{Key: client.KeyA, Action: client.Action{7, 0, 5}},
		{Key: client.KeyB, Action: client.Action{7, 0, 6}},
	})
}

func (w *World) prepareChanges(changes []client.BindingChange) {
	w.Changes = append([]client.BindingChange(nil), changes...)
	w.OriginalChanges = append([]client.BindingChange(nil), changes...)
	for _, change := range changes {
		w.Keyboard.Exchanges = append(w.Keyboard.Exchanges, keyboard.Exchange{
			Write: coreBindPacket(change), Response: []byte{0x54, 0xe4, 8},
		})
	}
}

func (w *World) theKeyboardHasBeenDisconnected() { w.Keyboard.Connected = false }

func (w *World) iRequestTheConnectedDevices() error {
	w.Devices, w.Err = w.Keyboard.Keyboard.Scan(context.Background())
	return w.Err
}

func (w *World) iRequestItsDeviceInformation() error {
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		var err error
		w.Info, err = k.Info(context.Background())
		return err
	})
	return w.Err
}

func (w *World) iRequestItsFirmwareVersion() {
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		_, err := k.FirmwareVersion(context.Background())
		return err
	})
}

func (w *World) theKeyboardIsReconnected() { w.Keyboard.Connected = true }

func (w *World) iApplyTheConfiguration() {
	if w.Changes == nil {
		// A disconnected-device scenario still needs a valid operation to attempt.
		w.Changes = []client.BindingChange{{Key: client.KeyA, Action: client.ActionB()}}
	}
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		return k.ApplyBindings(context.Background(), w.Changes)
	})
}

func (w *World) theKeyboardShouldAppearInTheDeviceList() error {
	if w.Err != nil {
		return w.Err
	}
	if !reflect.DeepEqual(w.Devices, []client.DeviceInfo{w.Keyboard.Device}) {
		return fmt.Errorf("devices=%+v; expected %+v", w.Devices, w.Keyboard.Device)
	}
	return nil
}

func (w *World) itsModelShouldBeReported() error {
	if w.Info.Device != w.Keyboard.Device {
		return fmt.Errorf("device information=%+v; expected %+v", w.Info.Device, w.Keyboard.Device)
	}
	return nil
}

func (w *World) itsFirmwareVersionShouldBeReported() error { return godog.ErrPending }

func (w *World) firmwareQueryingShouldReportUnsupported() error {
	if !errors.Is(w.Err, client.ErrUnsupported) {
		return fmt.Errorf("firmware query returned %v", w.Err)
	}
	return nil
}

func (w *World) theDeviceShouldNoLongerBeAvailable() error {
	if err := w.iRequestTheConnectedDevices(); err != nil {
		return err
	}
	if len(w.Devices) != 0 {
		return fmt.Errorf("disconnected device still listed: %+v", w.Devices)
	}
	k, err := w.Keyboard.Keyboard.Connect(context.Background(), w.Keyboard.Device.ID)
	if k != nil {
		return errors.Join(fmt.Errorf("disconnected device opened"), k.Close())
	}
	if !errors.Is(err, client.ErrDeviceNotFound) {
		return fmt.Errorf("disconnected open returned %v", err)
	}
	return nil
}

func (w *World) theDeviceShouldBecomeAvailableAgain() error {
	if err := w.iRequestTheConnectedDevices(); err != nil {
		return err
	}
	if err := w.theKeyboardShouldAppearInTheDeviceList(); err != nil {
		return err
	}
	return w.iRequestItsDeviceInformation()
}

func (w *World) theConfigurationShouldBeWrittenToTheKeyboard() error {
	if w.Err != nil {
		return w.Err
	}
	if !reflect.DeepEqual(w.Changes, w.OriginalChanges) {
		return fmt.Errorf("configuration input was modified")
	}
	if err := w.Keyboard.AssertComplete(); err != nil {
		return err
	}
	for _, change := range w.OriginalChanges {
		w.Keyboard.Exchanges = append(w.Keyboard.Exchanges, keyboard.Exchange{
			Write: coreActionPacket(change.Key), Response: coreActionReply(change.Key, change.Action),
		})
	}
	err := w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, func(k client.Keyboard) error {
		for _, change := range w.OriginalChanges {
			got, err := k.Binding(context.Background(), change.Key)
			if err != nil {
				return err
			}
			state := client.BindingAction
			if change.Action == client.ActionDisabled() {
				state = client.BindingDisabled
			}
			want := client.Binding{Key: change.Key, State: state, Action: change.Action, Effective: change.Action}
			if got != want {
				return fmt.Errorf("binding=%+v; expected %+v", got, want)
			}
		}
		return nil
	})
	return errors.Join(err, w.Keyboard.AssertComplete())
}

func (w *World) theMissingDeviceErrorShouldBeReturnedWithoutWriting() error {
	if !errors.Is(w.Err, client.ErrDeviceNotFound) {
		return fmt.Errorf("missing device error=%v", w.Err)
	}
	if len(w.Keyboard.Writes) != 0 || w.Keyboard.Reads != 0 {
		return fmt.Errorf("disconnected operation performed I/O")
	}
	return nil
}
