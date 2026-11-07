package steps

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ColinEge/8bd-kbd-api/client"
	"github.com/cucumber/godog"

	"github.com/ColinEge/8bd-kbd-api/test/fixtures/keyboard"
)

type World struct {
	Keyboard *keyboard.KeyboardFixture

	Err             error
	Devices         []client.DeviceInfo
	Info            client.KeyboardInfo
	Changes         []client.BindingChange
	OriginalChanges []client.BindingChange
	SelectedKey     client.KeyID
	Bindings        []client.Binding
	SelectedAction  client.Action
	ExpectedAction  client.Action
	ReadAction      client.Action
	ActionMapped    bool
	Binding         client.Binding
	OpenedKeyboard  client.Keyboard
	Keys            []client.KeyInfo
	Actions         []client.ActionInfo
	CallbackErr     error
	CleanupErr      error
	TransportErr    error
	WaitingErr      error
	BlockedWrites   atomic.Int32
	BlockedCloses   atomic.Int32
	Elapsed         time.Duration
}

func InitializeScenario(sc *godog.ScenarioContext) {
	w := &World{
		Keyboard: &keyboard.KeyboardFixture{},
	}

	sc.Before(func(
		ctx context.Context,
		scenario *godog.Scenario,
	) (context.Context, error) {
		w.Err = nil
		w.Devices, w.Changes, w.OriginalChanges, w.Bindings = nil, nil, nil, nil
		w.Info = client.KeyboardInfo{}
		w.SelectedKey = client.KeyA
		w.Binding = client.Binding{}
		w.OpenedKeyboard = nil
		w.Keys, w.Actions = nil, nil
		w.CallbackErr, w.CleanupErr, w.TransportErr, w.WaitingErr = nil, nil, nil, nil
		w.BlockedWrites.Store(0)
		w.BlockedCloses.Store(0)
		w.Elapsed = 0
		return ctx, nil
	})

	registerCommonSteps(sc, w)
	registerCoreSteps(sc, w)
	registerDeviceSteps(sc, w)
	registerRebindingSteps(sc, w)
	registerAudioSteps(sc, w)
	registerMacrosSteps(sc, w)
	registerProfileSteps(sc, w)
	registerAPIBindingSteps(sc, w)
}
