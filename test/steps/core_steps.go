package steps

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/ColinEge/8bd-kbd-api/client"
	"github.com/ColinEge/8bd-kbd-api/test/fixtures/keyboard"
	"github.com/ColinEge/8bd-kbd-api/transport"
	"github.com/cucumber/godog"
)

func registerCoreSteps(sc *godog.ScenarioContext, w *World) {
	sc.Step(`^the API operation should succeed$`, w.theAPIOperationShouldSucceed)
	sc.Step(`^the API error should be "([^"]*)"$`, w.theAPIErrorShouldBe)
	sc.Step(`^no packets should be written$`, w.noPacketsShouldBeWritten)
	sc.Step(`^exactly (\d+) keyboard connections? should be closed$`, w.keyboardConnectionsShouldBeClosed)
	sc.Step(`^all scripted packets should be consumed$`, w.allScriptedPacketsShouldBeConsumed)
	sc.Step(`^I inspect and mutate copies of the public catalogs$`, w.iInspectAndMutateCopiesOfThePublicCatalogs)
	sc.Step(`^the catalogs should remain deterministic and unchanged$`, w.theCatalogsShouldRemainDeterministicAndUnchanged)
	sc.Step(`^a batch containing "([^"]*)"$`, w.aBatchContaining)
	sc.Step(`^a scripted batch with a "([^"]*)" result$`, w.aScriptedBatchWithResult)
	sc.Step(`^I apply the prepared batch$`, w.iApplyThePreparedBatch)
	sc.Step(`^the supplied batch should remain unchanged$`, w.theSuppliedBatchShouldRemainUnchanged)
	sc.Step(`^exactly the first binding should be acknowledged in the apply error$`, w.theFirstBindingShouldBeAcknowledgedInTheApplyError)
	sc.Step(`^a scripted read-back of "([^"]*)"$`, w.aScriptedReadBackOf)
	sc.Step(`^I read a binding through the public API$`, w.iReadABindingThroughThePublicAPI)
	sc.Step(`^its stored and effective actions should match "([^"]*)"$`, w.storedAndEffectiveActionsShouldMatch)
	sc.Step(`^no binding value should be returned$`, w.noBindingValueShouldBeReturned)
	sc.Step(`^a scripted override list of "([^"]*)"$`, w.aScriptedOverrideListOf)
	sc.Step(`^I list bindings through the public API$`, w.iListBindingsThroughThePublicAPI)
	sc.Step(`^the override list result should be "([^"]*)"$`, w.theOverrideListResultShouldBe)
	sc.Step(`^I bind with an already cancelled context$`, w.iBindWithAnAlreadyCancelledContext)
	sc.Step(`^I run a command with a blocked "([^"]*)"$`, w.iRunACommandWithBlockedIO)
	sc.Step(`^the blocked command should stop within its timeout budget$`, w.theBlockedCommandShouldStopWithinItsTimeoutBudget)
	sc.Step(`^a second binding waits behind a blocked write$`, w.aSecondBindingWaitsBehindABlockedWrite)
	sc.Step(`^the waiting caller should time out without interleaving packets$`, w.theWaitingCallerShouldTimeOutWithoutInterleavingPackets)
	sc.Step(`^the blocked caller should retain cancellation identity$`, w.theBlockedCallerShouldRetainCancellationIdentity)
	sc.Step(`^a keyboard callback ends with "([^"]*)"$`, w.aKeyboardCallbackEndsWith)
	sc.Step(`^the callback and cleanup error identities should be preserved$`, w.callbackAndCleanupErrorIdentitiesShouldBePreserved)
	sc.Step(`^I close the client with an open keyboard$`, w.iCloseTheClientWithAnOpenKeyboard)
	sc.Step(`^the client and keyboard should reject further operations$`, w.theClientAndKeyboardShouldRejectFurtherOperations)
	sc.Step(`^repeated close calls should not close resources again$`, w.repeatedCloseCallsShouldNotCloseResourcesAgain)
}

func (w *World) operateKeyboard(f func(client.Keyboard) error) error {
	w.Err = w.Keyboard.Keyboard.WithKeyboard(context.Background(), w.Keyboard.Device.ID, f)
	return nil
}

func (w *World) theAPIOperationShouldSucceed() error { return w.Err }

func (w *World) theAPIErrorShouldBe(kind string) error {
	causes := map[string]error{
		"unsupported": client.ErrUnsupported, "not found": client.ErrDeviceNotFound,
		"invalid key": client.ErrInvalidKey, "invalid action": client.ErrInvalidAction,
		"rejected": client.ErrDeviceRejected, "malformed": client.ErrMalformedResponse,
		"transport": w.TransportErr, "cancelled": context.Canceled, "deadline": context.DeadlineExceeded,
	}
	cause, ok := causes[kind]
	if !ok || cause == nil {
		return fmt.Errorf("unknown error expectation %q", kind)
	}
	if !errors.Is(w.Err, cause) {
		return fmt.Errorf("error=%v; expected errors.Is(%v)", w.Err, cause)
	}
	return nil
}

func (w *World) noPacketsShouldBeWritten() error {
	if len(w.Keyboard.Writes) != 0 || w.Keyboard.Reads != 0 {
		return fmt.Errorf("writes=%d reads=%d", len(w.Keyboard.Writes), w.Keyboard.Reads)
	}
	return w.Keyboard.AssertComplete()
}

func (w *World) keyboardConnectionsShouldBeClosed(n int) error {
	if w.Keyboard.SessionCloses != n {
		return fmt.Errorf("closes=%d; expected %d", w.Keyboard.SessionCloses, n)
	}
	return nil
}

func (w *World) allScriptedPacketsShouldBeConsumed() error {
	return w.Keyboard.AssertComplete()
}

func (w *World) iInspectAndMutateCopiesOfThePublicCatalogs() error {
	return w.operateKeyboard(func(k client.Keyboard) error {
		w.Keys, w.Actions = k.Keys(), k.Actions()
		if len(w.Keys) == 0 || len(w.Actions) == 0 {
			return fmt.Errorf("empty built-in catalog")
		}
		keyCopy, actionCopy := k.Keys(), k.Actions()
		keyCopy[0].Label, actionCopy[0].Label = "changed", "changed"
		if !reflect.DeepEqual(w.Keys, k.Keys()) || !reflect.DeepEqual(w.Actions, k.Actions()) {
			return fmt.Errorf("catalog was mutated")
		}
		return nil
	})
}

func (w *World) theCatalogsShouldRemainDeterministicAndUnchanged() error {
	if w.Err != nil {
		return w.Err
	}
	seenKeys, seenActions := map[client.KeyID]bool{}, map[client.Action]bool{}
	for _, key := range w.Keys {
		if seenKeys[key.ID] {
			return fmt.Errorf("duplicate key %s", key.ID)
		}
		seenKeys[key.ID] = true
	}
	for _, action := range w.Actions {
		if seenActions[action.Action] {
			return fmt.Errorf("duplicate action %s", action.Action)
		}
		seenActions[action.Action] = true
	}
	if !seenKeys[client.KeyA] || !seenActions[client.ActionF13()] || !seenActions[client.ActionVolumeUp()] {
		return fmt.Errorf("required catalog entries missing")
	}
	return nil
}

func (w *World) aBatchContaining(kind string) error {
	w.Changes = []client.BindingChange{{Key: client.KeyA, Action: client.ActionB()}, {Key: client.KeyB, Action: client.ActionC()}}
	switch kind {
	case "invalid key":
		w.Changes[0].Key = 0
	case "zero action":
		w.Changes[0].Action = client.Action{}
	case "unsupported action":
		w.Changes[0].Action = client.Action{255, 1, 2}
	case "invalid second key":
		w.Changes[1].Key = 255
	case "duplicate key":
		w.Changes[1].Key = w.Changes[0].Key
	default:
		return fmt.Errorf("unknown batch %q", kind)
	}
	w.OriginalChanges = append([]client.BindingChange(nil), w.Changes...)
	return nil
}

func (w *World) aScriptedBatchWithResult(result string) error {
	w.prepareChanges([]client.BindingChange{{Key: client.KeyA, Action: client.ActionB()}, {Key: client.KeyB, Action: client.ActionC()}})
	w.TransportErr = errors.New("scripted batch transport failure")
	switch result {
	case "success":
	case "rejection":
		w.Keyboard.Exchanges[1].Response = []byte{0x54, 0xe4, 9}
	case "write failure":
		w.Keyboard.Exchanges[1].WriteErr = w.TransportErr
	case "read failure":
		w.Keyboard.Exchanges[1].ReadErr = w.TransportErr
	default:
		return fmt.Errorf("unknown result %q", result)
	}
	return nil
}

func (w *World) iApplyThePreparedBatch() error {
	return w.operateKeyboard(func(k client.Keyboard) error {
		return k.ApplyBindings(context.Background(), w.Changes)
	})
}

func (w *World) theSuppliedBatchShouldRemainUnchanged() error {
	if !reflect.DeepEqual(w.Changes, w.OriginalChanges) {
		return fmt.Errorf("batch input changed: %+v", w.Changes)
	}
	return nil
}

func (w *World) theFirstBindingShouldBeAcknowledgedInTheApplyError() error {
	var applyErr *client.ApplyError
	if !errors.As(w.Err, &applyErr) || !reflect.DeepEqual(applyErr.Completed, w.Changes[:1]) || applyErr.Failed != w.Changes[1] {
		return fmt.Errorf("incorrect partial application error: %v", w.Err)
	}
	return nil
}

func expectedCoreBinding(state string) (client.Binding, error) {
	b := client.Binding{Key: client.KeyA}
	switch state {
	case "default":
		b.State, b.Effective = client.BindingDefault, client.ActionA()
	case "disabled":
		b.State, b.Action, b.Effective = client.BindingDisabled, client.ActionDisabled(), client.ActionDisabled()
	case "action":
		b.State, b.Action, b.Effective = client.BindingAction, client.ActionB(), client.ActionB()
	default:
		return b, fmt.Errorf("unknown binding state %q", state)
	}
	return b, nil
}

func (w *World) aScriptedReadBackOf(kind string) error {
	response := coreActionReply(client.KeyA, client.ActionB())
	switch kind {
	case "default", "disabled", "action":
		b, err := expectedCoreBinding(kind)
		if err != nil {
			return err
		}
		response = coreActionReply(client.KeyA, b.Action)
	case "truncated":
		response = []byte{0x54}
	case "wrong key":
		response[2] = byte(client.KeyB)
	case "wrong header":
		response[0] = 0x52
	case "oversized":
		response = append(response, 0)
	default:
		return fmt.Errorf("unknown reply %q", kind)
	}
	w.Keyboard.Exchanges = []keyboard.Exchange{{Write: coreActionPacket(client.KeyA), Response: response}}
	return nil
}

func (w *World) iReadABindingThroughThePublicAPI() error {
	return w.operateKeyboard(func(k client.Keyboard) error {
		var err error
		w.Binding, err = k.Binding(context.Background(), client.KeyA)
		return err
	})
}

func (w *World) storedAndEffectiveActionsShouldMatch(state string) error {
	want, err := expectedCoreBinding(state)
	if err != nil {
		return err
	}
	if w.Err != nil || w.Binding != want {
		return fmt.Errorf("binding=%+v error=%v; expected %+v", w.Binding, w.Err, want)
	}
	return nil
}

func (w *World) noBindingValueShouldBeReturned() error {
	if w.Binding != (client.Binding{}) {
		return fmt.Errorf("malformed response returned %+v", w.Binding)
	}
	return nil
}

func (w *World) aScriptedOverrideListOf(kind string) error {
	request, response := make([]byte, 33), make([]byte, 33)
	copy(request, []byte{0x52, 0x81})
	copy(response, []byte{0x54, 0x81})
	switch kind {
	case "empty":
	case "mapped":
		copy(response[2:], []byte{4, 7})
	case "malformed":
		response = []byte{0x54}
	default:
		return fmt.Errorf("unknown list %q", kind)
	}
	w.Keyboard.Exchanges = []keyboard.Exchange{{Write: request, Response: response}}
	if kind == "mapped" {
		w.Keyboard.Exchanges = append(w.Keyboard.Exchanges, keyboard.Exchange{Write: coreActionPacket(client.KeyA), Response: coreActionReply(client.KeyA, client.ActionB())})
	}
	return nil
}

func (w *World) iListBindingsThroughThePublicAPI() error {
	return w.operateKeyboard(func(k client.Keyboard) error {
		var err error
		w.Bindings, err = k.Bindings(context.Background())
		return err
	})
}

func (w *World) theOverrideListResultShouldBe(kind string) error {
	if kind == "malformed" {
		if !errors.Is(w.Err, client.ErrMalformedResponse) || w.Bindings != nil {
			return fmt.Errorf("malformed list returned %+v, %v", w.Bindings, w.Err)
		}
		return nil
	}
	want := []client.Binding{}
	if kind == "mapped" {
		b, err := expectedCoreBinding("action")
		if err != nil {
			return err
		}
		want = append(want, b)
	}
	if w.Err != nil || !reflect.DeepEqual(w.Bindings, want) {
		return fmt.Errorf("bindings=%+v error=%v; expected %+v", w.Bindings, w.Err, want)
	}
	return nil
}

func (w *World) iBindWithAnAlreadyCancelledContext() error {
	return w.operateKeyboard(func(k client.Keyboard) error {
		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		return k.Bind(cancelled, client.KeyA, client.ActionB())
	})
}

func (w *World) iRunACommandWithBlockedIO(direction string) (err error) {
	if direction != "read" && direction != "write" {
		return fmt.Errorf("unknown blocked direction %q", direction)
	}
	w.BlockedWrites.Store(0)
	w.BlockedCloses.Store(0)
	mock := &transport.Mock{
		WriteFunc: func(ctx context.Context, data []byte) error {
			w.BlockedWrites.Add(1)
			if direction == "write" {
				<-ctx.Done()
				return ctx.Err()
			}
			return nil
		},
		ReadFunc:  func(ctx context.Context) ([]byte, error) { <-ctx.Done(); return nil, ctx.Err() },
		CloseFunc: func() error { w.BlockedCloses.Add(1); return nil },
	}
	c := client.New(client.WithTimeout(30*time.Millisecond), client.WithDeviceProvider(&blockingProvider{device: w.Keyboard.Device, transport: mock}))
	defer func() { err = errors.Join(err, c.Close()) }()
	parent, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	w.Err = c.WithKeyboard(parent, w.Keyboard.Device.ID, func(k client.Keyboard) error {
		return k.Bind(parent, client.KeyA, client.ActionB())
	})
	w.Elapsed = time.Since(start)
	return nil
}

func (w *World) theBlockedCommandShouldStopWithinItsTimeoutBudget() error {
	if w.Elapsed > 500*time.Millisecond || w.BlockedWrites.Load() != 1 || w.BlockedCloses.Load() != 1 {
		return fmt.Errorf("timeout elapsed=%v writes=%d closes=%d", w.Elapsed, w.BlockedWrites.Load(), w.BlockedCloses.Load())
	}
	return nil
}

func (w *World) aSecondBindingWaitsBehindABlockedWrite() (err error) {
	w.BlockedWrites.Store(0)
	w.BlockedCloses.Store(0)
	started := make(chan struct{}, 1)
	c := client.New(client.WithDeviceProvider(&blockingProvider{device: w.Keyboard.Device, transport: &transport.Mock{
		WriteFunc: func(ctx context.Context, data []byte) error {
			w.BlockedWrites.Add(1)
			started <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		},
		CloseFunc: func() error { w.BlockedCloses.Add(1); return nil },
	}}))
	defer func() { err = errors.Join(err, c.Close()) }()
	ctx := context.Background()
	k, err := c.Connect(ctx, w.Keyboard.Device.ID)
	if err != nil {
		return err
	}
	firstCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- k.Bind(firstCtx, client.KeyA, client.ActionB()) }()
	select {
	case <-started:
	case <-firstCtx.Done():
		return fmt.Errorf("first command did not start: %w", firstCtx.Err())
	}
	waitCtx, stop := context.WithTimeout(ctx, 30*time.Millisecond)
	w.WaitingErr = k.Bind(waitCtx, client.KeyB, client.ActionC())
	stop()
	cancel()
	w.Err = <-done
	return nil
}

func (w *World) theWaitingCallerShouldTimeOutWithoutInterleavingPackets() error {
	if !errors.Is(w.WaitingErr, context.DeadlineExceeded) || w.BlockedWrites.Load() != 1 || w.BlockedCloses.Load() != 1 {
		return fmt.Errorf("waiting error=%v writes=%d closes=%d", w.WaitingErr, w.BlockedWrites.Load(), w.BlockedCloses.Load())
	}
	return nil
}

func (w *World) theBlockedCallerShouldRetainCancellationIdentity() error {
	if !errors.Is(w.Err, context.Canceled) {
		return fmt.Errorf("blocked caller lost cancellation: %v", w.Err)
	}
	return nil
}

func (w *World) aKeyboardCallbackEndsWith(outcome string) error {
	w.CallbackErr, w.CleanupErr = nil, nil
	switch outcome {
	case "callback error", "panic":
		w.CallbackErr = errors.New("callback failure")
	case "cleanup error":
		w.CleanupErr = errors.New("cleanup failure")
	case "callback+cleanup":
		w.CallbackErr, w.CleanupErr = errors.New("callback failure"), errors.New("cleanup failure")
	default:
		return fmt.Errorf("unknown callback outcome %q", outcome)
	}
	w.Keyboard.SessionCloseErr = w.CleanupErr
	return w.operateKeyboard(func(client.Keyboard) error {
		if outcome == "panic" {
			panic(w.CallbackErr)
		}
		return w.CallbackErr
	})
}

func (w *World) callbackAndCleanupErrorIdentitiesShouldBePreserved() error {
	for _, cause := range []error{w.CallbackErr, w.CleanupErr} {
		if cause != nil && !errors.Is(w.Err, cause) {
			return fmt.Errorf("error %v lost cause %v", w.Err, cause)
		}
	}
	return w.theOperationShouldFail()
}

func (w *World) iCloseTheClientWithAnOpenKeyboard() error {
	var err error
	w.OpenedKeyboard, err = w.Keyboard.Keyboard.Connect(context.Background(), w.Keyboard.Device.ID)
	if err != nil {
		return err
	}
	return w.Keyboard.Keyboard.Close()
}

func (w *World) theClientAndKeyboardShouldRejectFurtherOperations() error {
	ctx := context.Background()
	_, scanErr := w.Keyboard.Keyboard.Scan(ctx)
	_, openErr := w.Keyboard.Keyboard.Connect(ctx, w.Keyboard.Device.ID)
	_, infoErr := w.OpenedKeyboard.Info(ctx)
	bindErr := w.OpenedKeyboard.Bind(ctx, client.KeyA, client.ActionB())
	for _, err := range []error{scanErr, openErr, infoErr, bindErr} {
		if !errors.Is(err, client.ErrDeviceClosed) {
			return fmt.Errorf("closed operation returned %v", err)
		}
	}
	return nil
}

func (w *World) repeatedCloseCallsShouldNotCloseResourcesAgain() error {
	if err := errors.Join(w.OpenedKeyboard.Close(), w.Keyboard.Keyboard.Close()); err != nil {
		return err
	}
	if w.Keyboard.SessionCloses != 1 || w.Keyboard.ProviderCloses != 1 {
		return fmt.Errorf("keyboard closes=%d provider closes=%d", w.Keyboard.SessionCloses, w.Keyboard.ProviderCloses)
	}
	return nil
}

func coreBindPacket(change client.BindingChange) []byte {
	packet := make([]byte, 33)
	copy(packet, []byte{0x52, 0xfa, 3, 0x0c, 0, 0xaa, 9, 0x71, byte(change.Key)})
	copy(packet[9:], change.Action[:])
	return packet
}

func coreActionPacket(key client.KeyID) []byte {
	packet := make([]byte, 33)
	copy(packet, []byte{0x52, 0x83, byte(key)})
	return packet
}

func coreActionReply(key client.KeyID, action client.Action) []byte {
	packet := make([]byte, 33)
	copy(packet, []byte{0x54, 0x83, byte(key)})
	copy(packet[3:], action[:])
	return packet
}

type blockingProvider struct {
	device    client.DeviceInfo
	transport transport.Transport
}

func (p *blockingProvider) ListDevices(ctx context.Context) ([]client.DeviceInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []client.DeviceInfo{p.device}, nil
}

func (p *blockingProvider) OpenDevice(ctx context.Context, id client.DeviceID) (*client.OpenedDevice, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id != p.device.ID {
		return nil, client.ErrDeviceNotFound
	}
	return &client.OpenedDevice{Info: p.device, Transport: p.transport}, nil
}

func (p *blockingProvider) Close() error { return nil }
