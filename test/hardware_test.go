package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ColinEge/8bd-kbd-api/client"
	"github.com/ColinEge/8bd-kbd-api/keydefs"
	"github.com/ColinEge/8bd-kbd-api/test/fixtures/keyboard"
	"github.com/cucumber/godog"
)

func TestHardwareReadOnlyFeatures(t *testing.T) {
	if os.Getenv("TEST_8BITDO_ACCEPTANCE") != "1" {
		t.Skip("set TEST_8BITDO_ACCEPTANCE=1 plus explicit device and key to run read-only hardware features")
	}
	runHardwareFeatures(t, false)
}

func TestHardwareWriteFeatures(t *testing.T) {
	if os.Getenv("TEST_8BITDO_ACCEPTANCE_WRITE") != "1" {
		t.Skip("set TEST_8BITDO_ACCEPTANCE_WRITE=1 plus explicit device and spare key to run hardware writes")
	}
	runHardwareFeatures(t, true)
}

func hardwareSelection(device, key string) (client.DeviceID, client.KeyID, error) {
	if device == "" || strings.TrimSpace(device) != device {
		return "", 0, fmt.Errorf("TEST_8BITDO_DEVICE must be an explicit scan ID")
	}
	base := 10
	if strings.HasPrefix(key, "0x") || strings.HasPrefix(key, "0X") {
		base = 0
	}
	value, err := strconv.ParseUint(key, base, 8)
	if err != nil {
		return "", 0, fmt.Errorf("TEST_8BITDO_KEY must be an explicit firmware key ID: %w", err)
	}
	// Modifier defaults are not hardware-verified. Use an ordinary spare key.
	if value < 4 || value > 0x63 {
		return "", 0, fmt.Errorf("hardware features require an ordinary firmware key ID from 0x04 to 0x63")
	}
	return client.DeviceID(device), client.KeyID(value), nil
}

func hardwareAction(name string) (client.Action, error) {
	switch name {
	case "B":
		return client.ActionB(), nil
	case "Ctrl+C":
		return client.Chord(client.LeftCtrl, client.ActionC())
	case "F13":
		return client.ActionF13(), nil
	case "disabled":
		return client.ActionDisabled(), nil
	case "volume up":
		return client.ActionVolumeUp(), nil
	case "volume down":
		return client.ActionVolumeDown(), nil
	case "mute":
		return client.ActionMute(), nil
	default:
		return client.Action{}, fmt.Errorf("unknown hardware action %q", name)
	}
}

// Expectations are independent of the helpers used to generate the write.
func expectedHardwareAction(name string) (client.Action, error) {
	switch name {
	case "B":
		return client.Action{7, 0, 5}, nil
	case "Ctrl+C":
		return client.Action{7, 0xe0, 6}, nil
	case "F13":
		return client.Action{7, 0, 0x68}, nil
	case "disabled":
		return client.Action{7, 0, 0}, nil
	case "volume up":
		return client.Action{0x0c, 0xe9, 0}, nil
	case "volume down":
		return client.Action{0x0c, 0xea, 0}, nil
	case "mute":
		return client.Action{0x0c, 0xe2, 0}, nil
	default:
		return client.Action{}, fmt.Errorf("unknown expected action %q", name)
	}
}

func runHardwareFeatures(t *testing.T, writes bool) {
	t.Helper()
	id, key, err := hardwareSelection(os.Getenv("TEST_8BITDO_DEVICE"), os.Getenv("TEST_8BITDO_KEY"))
	if err != nil {
		t.Fatal(err)
	}
	path := "./features/hardware_readonly.feature"
	if writes {
		path = "./features/hardware_writes.feature"
	}
	suite := godog.TestSuite{
		Name: "physical-keyboard-api",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var c client.Client
			var saved, first, second client.Binding
			var discovered client.DeviceInfo
			var info client.KeyboardInfo
			var restoreNeeded bool
			var scenarioCtx context.Context
			var cancel context.CancelFunc

			readBinding := func(ctx context.Context, c client.Client) (client.Binding, error) {
				var binding client.Binding
				err := c.WithKeyboard(ctx, id, func(k client.Keyboard) error {
					var err error
					binding, err = k.Binding(ctx, key)
					return err
				})
				return binding, err
			}
			sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
				saved, first, second = client.Binding{}, client.Binding{}, client.Binding{}
				discovered, info = client.DeviceInfo{}, client.KeyboardInfo{}
				restoreNeeded = false
				scenarioCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
				c = client.New(client.WithTimeout(5 * time.Second))
				if !writes {
					return scenarioCtx, nil
				}
				var err error
				saved, err = readBinding(scenarioCtx, c)
				if err != nil {
					return scenarioCtx, err
				}
				if saved.State != client.BindingDefault && !keydefs.SupportedAction(saved.Action) {
					return scenarioCtx, fmt.Errorf("test key %s has an unsupported stored action %s; cannot safely restore it", key, saved.Action)
				}
				return scenarioCtx, nil
			})
			sc.After(func(ctx context.Context, scenario *godog.Scenario, scenarioErr error) (context.Context, error) {
				defer cancel()
				var err error
				if restoreNeeded {
					// Independent cleanup context remains usable after scenario timeout.
					cleanupCtx, stop := context.WithTimeout(context.Background(), 15*time.Second)
					cleanupClient := client.New(client.WithTimeout(5 * time.Second))
					err = restoreHardwareBinding(cleanupCtx, cleanupClient, id, saved)
					err = errors.Join(err, cleanupClient.Close())
					stop()
					if err != nil {
						err = fmt.Errorf("RESTORATION FAILED for device %s key %s; original action=%s: %w", id, key, saved.Action, err)
					}
				}
				return ctx, errors.Join(err, c.Close())
			})
			sc.Step(`^I scan and open the selected physical keyboard$`, func() error {
				devices, err := c.Scan(scenarioCtx)
				if err != nil {
					return err
				}
				found := false
				for _, device := range devices {
					if device.ID == id {
						discovered, found = device, true
						break
					}
				}
				if !found {
					return fmt.Errorf("selected device %s not in scan", id)
				}
				return c.WithKeyboard(scenarioCtx, id, func(k client.Keyboard) error { var err error; info, err = k.Info(scenarioCtx); return err })
			})
			sc.Step(`^its connection information should match discovery$`, func() error {
				if info.Device != discovered {
					return fmt.Errorf("info=%+v; discovery=%+v", info.Device, discovered)
				}
				return nil
			})
			sc.Step(`^firmware querying should explicitly report unsupported$`, func() error {
				return c.WithKeyboard(scenarioCtx, id, func(k client.Keyboard) error {
					version, err := k.FirmwareVersion(scenarioCtx)
					if !errors.Is(err, client.ErrUnsupported) || version != "" {
						return fmt.Errorf("firmware returned %q, %v", version, err)
					}
					return nil
				})
			})
			sc.Step(`^I read the selected key twice through separate connections$`, func() error {
				var err error
				first, err = readBinding(scenarioCtx, c)
				if err != nil {
					return err
				}
				second, err = readBinding(scenarioCtx, c)
				return err
			})
			sc.Step(`^the physical binding state and actions should be consistent$`, func() error {
				if first != second || first.Key != key {
					return fmt.Errorf("inconsistent bindings: %+v and %+v", first, second)
				}
				switch first.State {
				case client.BindingDefault:
					if first.Action != (client.Action{}) || first.Effective != (client.Action{7, 0, byte(key)}) {
						return fmt.Errorf("invalid default binding: %+v", first)
					}
				case client.BindingAction:
					if first.Action == (client.Action{}) || first.Action == (client.Action{7, 0, 0}) || first.Effective != first.Action {
						return fmt.Errorf("invalid action binding: %+v", first)
					}
				case client.BindingDisabled:
					if first.Action != (client.Action{7, 0, 0}) || first.Effective != first.Action {
						return fmt.Errorf("invalid disabled binding: %+v", first)
					}
				default:
					return fmt.Errorf("unknown binding state %q", first.State)
				}
				t.Logf("device=%s key=%s binding=%+v", id, key, first)
				return nil
			})
			sc.Step(`^I bind the selected physical key to "([^"]*)"$`, func(name string) error {
				action, err := hardwareAction(name)
				if err != nil {
					return err
				}
				restoreNeeded = true
				return c.WithKeyboard(scenarioCtx, id, func(k client.Keyboard) error {
					if name == "disabled" {
						return k.Disable(scenarioCtx, key)
					}
					return k.Bind(scenarioCtx, key, action)
				})
			})
			sc.Step(`^physical read-back should report "([^"]*)"$`, func(name string) error {
				expected, err := expectedHardwareAction(name)
				if err != nil {
					return err
				}
				binding, err := readBinding(scenarioCtx, c)
				if err != nil {
					return err
				}
				state := client.BindingAction
				if name == "disabled" {
					state = client.BindingDisabled
				}
				want := client.Binding{Key: key, State: state, Action: expected, Effective: expected}
				if binding != want {
					return fmt.Errorf("physical binding=%+v; expected %+v", binding, want)
				}
				return nil
			})
			sc.Step(`^I restore the selected physical key$`, func() error {
				restoreNeeded = true
				return c.WithKeyboard(scenarioCtx, id, func(k client.Keyboard) error { return k.Restore(scenarioCtx, key) })
			})
			sc.Step(`^its effective action should be the factory default$`, func() error {
				binding, err := readBinding(scenarioCtx, c)
				if err != nil {
					return err
				}
				expected := client.Action{7, 0, byte(key)}
				if binding.Effective != expected {
					return fmt.Errorf("effective action=%s; expected %s", binding.Effective, expected)
				}
				return nil
			})
			sc.Step(`^I recreate the API client$`, func() error {
				if err := c.Close(); err != nil {
					return err
				}
				c = client.New(client.WithTimeout(5 * time.Second))
				return nil
			})
		},
		Options: &godog.Options{Format: "pretty", Paths: []string{path}, Strict: true, TestingT: t},
	}
	if status := suite.Run(); status != 0 {
		t.Fatalf("hardware features failed with status %d", status)
	}
}

func restoreHardwareBinding(ctx context.Context, c client.Client, id client.DeviceID, saved client.Binding) error {
	return c.WithKeyboard(ctx, id, func(k client.Keyboard) error {
		var restoreErr error
		if saved.State == client.BindingDefault {
			restoreErr = k.Restore(ctx, saved.Key)
		} else {
			restoreErr = k.Bind(ctx, saved.Key, saved.Action)
		}
		if restoreErr != nil {
			return restoreErr
		}
		actual, err := k.Binding(ctx, saved.Key)
		if err != nil {
			return err
		}
		matches := actual == saved
		if saved.State == client.BindingDefault {
			matches = actual.Key == saved.Key && actual.Effective == saved.Effective &&
				(actual.State == client.BindingDefault ||
					(actual.State == client.BindingAction && actual.Action == saved.Effective))
		}
		if !matches {
			return fmt.Errorf("restored binding=%+v; expected %+v", actual, saved)
		}
		return nil
	})
}

func TestHardwareRestoration(t *testing.T) {
	for _, kind := range []string{"success", "rejected", "wrong read-back", "write failure", "default absent", "default explicit", "default wrong read-back", "disabled"} {
		t.Run(kind, func(t *testing.T) {
			f := &keyboard.KeyboardFixture{}
			if err := f.Reset(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := f.Close(); err != nil {
					t.Error(err)
				}
			})
			saved := client.Binding{Key: client.KeyA, State: client.BindingAction, Action: client.Action{7, 0, 6}, Effective: client.Action{7, 0, 6}}
			write := make([]byte, 33)
			copy(write, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, 4, 7, 0, 6})
			query, reply := make([]byte, 33), make([]byte, 33)
			copy(query, []byte{0x52, 0x83, 4})
			copy(reply, []byte{0x54, 0x83, 4, 7, 0, 6})
			switch kind {
			case "default absent", "default explicit", "default wrong read-back":
				saved = client.Binding{Key: client.KeyA, State: client.BindingDefault, Effective: client.ActionA()}
				write[11] = 4
				copy(reply[3:], make([]byte, 6))
				if kind == "default explicit" {
					copy(reply[3:], []byte{7, 0, 4})
				} else if kind == "default wrong read-back" {
					copy(reply[3:], []byte{7, 0, 5})
				}
			case "disabled":
				saved = client.Binding{Key: client.KeyA, State: client.BindingDisabled, Action: client.ActionDisabled(), Effective: client.ActionDisabled()}
				write[11], reply[5] = 0, 0
			}
			cause := errors.New("scripted cleanup write failure")
			f.Exchanges = []keyboard.Exchange{{Write: write, Response: []byte{0x54, 0xe4, 8}}}
			switch kind {
			case "rejected":
				f.Exchanges[0].Response = []byte{0x54, 0xe4, 9}
			case "write failure":
				f.Exchanges[0].WriteErr = cause
			case "wrong read-back":
				reply[5] = 5
			}
			if kind != "rejected" && kind != "write failure" {
				f.Exchanges = append(f.Exchanges, keyboard.Exchange{Write: query, Response: reply})
			}
			err := restoreHardwareBinding(context.Background(), f.Keyboard, f.Device.ID, saved)
			switch kind {
			case "success", "default absent", "default explicit", "disabled":
				if err != nil {
					t.Fatal(err)
				}
			case "rejected":
				if !errors.Is(err, client.ErrDeviceRejected) {
					t.Fatalf("rejection lost: %v", err)
				}
			case "write failure":
				if !errors.Is(err, cause) {
					t.Fatalf("write cause lost: %v", err)
				}
			case "wrong read-back", "default wrong read-back":
				if err == nil {
					t.Fatal("incorrect restoration accepted")
				}
			}
			if f.SessionCloses != 1 {
				t.Fatalf("cleanup connections closed=%d", f.SessionCloses)
			}
			if err := f.AssertComplete(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHardwareSelection(t *testing.T) {
	for _, test := range []struct{ device, key string }{
		{"", "4"}, {" 001-003", "4"}, {"001-003", ""}, {"001-003", "A"}, {"001-003", "0"},
		{"001-003", "0x64"}, {"001-003", "0x65"}, {"001-003", "0x66"}, {"001-003", "256"},
	} {
		if _, _, err := hardwareSelection(test.device, test.key); err == nil {
			t.Fatalf("accepted invalid hardware selection %+v", test)
		}
	}
	for _, key := range []string{"4", "0x04", "0X04"} {
		id, got, err := hardwareSelection("001-003", key)
		if err != nil || id != "001-003" || got != client.KeyA {
			t.Fatalf("selection=%s/%s error=%v", id, got, err)
		}
	}
	for _, name := range []string{"B", "Ctrl+C", "F13", "disabled", "volume up", "volume down", "mute"} {
		got, err := hardwareAction(name)
		want, wantErr := expectedHardwareAction(name)
		if err != nil || wantErr != nil || got != want {
			t.Fatalf("action %q=%s; expected %s; errors=%v/%v", name, got, want, err, wantErr)
		}
	}
}
