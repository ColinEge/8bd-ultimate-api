package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ColinEge/8bd-ultimate-api/internal/command"
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
	"github.com/ColinEge/8bd-ultimate-api/transport"
)

type keyboard struct {
	info      KeyboardInfo
	transport transport.Transport
	timeout   time.Duration
	gate      chan struct{}
	closed    bool
}

func (s *keyboard) Info(ctx context.Context) (KeyboardInfo, error) {
	if err := s.acquire(ctx); err != nil {
		return KeyboardInfo{}, err
	}
	defer s.release()
	return s.info, nil
}

func (s *keyboard) FirmwareVersion(ctx context.Context) (string, error) {
	if err := s.acquire(ctx); err != nil {
		return "", err
	}
	defer s.release()
	return "", fmt.Errorf("%w: firmware version query is not verified", ErrUnsupported)
}

func (s *keyboard) acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.gate:
	}
	if s.closed {
		s.release()
		return ErrDeviceClosed
	}
	if err := ctx.Err(); err != nil {
		s.release()
		return err
	}
	return nil
}

func (s *keyboard) release() { s.gate <- struct{}{} }

func validateKey(key KeyID) error {
	// These are firmware IDs, not a guarantee of a model's physical layout.
	if key < 4 || key > 0x6b {
		return fmt.Errorf("%w: firmware key %s", ErrInvalidKey, key)
	}
	return nil
}

func validateChange(change BindingChange) error {
	if err := validateKey(change.Key); err != nil {
		return err
	}
	if change.Action == (Action{}) {
		return fmt.Errorf("%w: zero is an absent override, not a bindable action", ErrInvalidAction)
	}
	if !keydefs.SupportedAction(change.Action) {
		return fmt.Errorf("%w: action %s", ErrUnsupported, change.Action)
	}
	return nil
}

func (s *keyboard) Keys() []KeyInfo {
	keys := make([]KeyInfo, 0)
	seen := make(map[KeyID]bool)
	for _, entry := range keydefs.Catalog() {
		key, ok := keydefs.ActionToPhysicalKey(entry.Action)
		if !ok {
			continue
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		keys = append(keys, KeyInfo{ID: key, Label: entry.KeyLabel, Default: entry.Action})
	}
	return keys
}

func (s *keyboard) Actions() []ActionInfo {
	entries := keydefs.Catalog()
	actions := make([]ActionInfo, 0, len(entries))
	seen := make(map[Action]bool)
	for _, entry := range entries {
		if entry.Category == keydefs.CategoryKeyOnly {
			continue
		}
		if seen[entry.Action] {
			continue
		}
		seen[entry.Action] = true
		actions = append(actions, ActionInfo{Action: entry.Action, Label: entry.Label, Category: entry.Category})
	}
	return actions
}

func (s *keyboard) Binding(ctx context.Context, key KeyID) (Binding, error) {
	if err := s.acquire(ctx); err != nil {
		return Binding{}, err
	}
	defer s.release()
	return s.binding(ctx, key)
}

func (s *keyboard) binding(ctx context.Context, key KeyID) (Binding, error) {
	if err := validateKey(key); err != nil {
		return Binding{}, err
	}
	r := command.KeyActionResponse{}
	if err := s.sendResponse(ctx, &command.GetKeyAction{PhysicalKey: key}, &r); err != nil {
		return Binding{}, err
	}
	action, err := r.Decode()
	if err != nil {
		return Binding{}, fmt.Errorf("%w: read binding for key %s: %w", ErrMalformedResponse, key, err)
	}
	if r.Payload[2] != byte(key) {
		return Binding{}, fmt.Errorf("%w: response key %s does not match %s", ErrMalformedResponse, KeyID(r.Payload[2]), key)
	}
	binding := Binding{Key: key, Action: action, Effective: action, State: BindingAction}
	switch action {
	case Action{}:
		binding.State = BindingDefault
		binding.Effective = keydefs.PhysicalKeyToAction(byte(key))
	case ActionDisabled():
		binding.State = BindingDisabled
	}
	return binding, nil
}

// Bindings returns stored overrides, not every physical key on the device.
func (s *keyboard) Bindings(ctx context.Context) ([]Binding, error) {
	if err := s.acquire(ctx); err != nil {
		return nil, err
	}
	defer s.release()
	var response command.RemappedKeysResponse
	if err := s.sendResponse(ctx, &command.GetRemappedKeys{}, &response); err != nil {
		return nil, err
	}
	keys, err := response.Decode()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedResponse, err)
	}
	bindings := make([]Binding, 0, len(keys))
	seen := make(map[KeyID]bool)
	for _, key := range keys {
		if err := validateKey(key.PhysicalKey); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrMalformedResponse, err)
		}
		if seen[key.PhysicalKey] {
			return nil, fmt.Errorf("%w: duplicate mapped key %s", ErrMalformedResponse, key.PhysicalKey)
		}
		seen[key.PhysicalKey] = true
		binding, err := s.binding(ctx, key.PhysicalKey)
		if err != nil {
			return nil, err
		}
		if binding.State == BindingDefault {
			return nil, fmt.Errorf("%w: listed key %s has no override", ErrMalformedResponse, binding.Key)
		}
		bindings = append(bindings, binding)
	}
	return bindings, nil
}

func (s *keyboard) Bind(ctx context.Context, key KeyID, action Action) error {
	if err := s.acquire(ctx); err != nil {
		return err
	}
	defer s.release()
	if err := validateChange(BindingChange{Key: key, Action: action}); err != nil {
		return err
	}
	return s.bind(ctx, key, action)
}

func (s *keyboard) bind(ctx context.Context, key KeyID, action Action) error {
	var response command.StatusResponse
	if err := s.sendResponse(ctx, &command.BindAction{Key: key, ActionCode: action}, &response); err != nil {
		return err
	}
	status, err := response.DecodeStatus()
	if err != nil {
		return fmt.Errorf("%w: bind action failed: %w", ErrMalformedResponse, err)
	}
	if status != command.StatusResponseOk {
		return fmt.Errorf("%w: bind action failed: device responded %s", ErrDeviceRejected, status)
	}
	return nil
}

func (s *keyboard) Disable(ctx context.Context, key KeyID) error {
	return s.Bind(ctx, key, ActionDisabled())
}

// ResetToDefault binds the key's factory-default action, not a previous custom mapping.
func (s *keyboard) ResetToDefault(ctx context.Context, key KeyID) error {
	if err := s.Bind(ctx, key, keydefs.PhysicalKeyToAction(byte(key))); err != nil {
		return fmt.Errorf("reset key %s to default failed: %w", key, err)
	}
	return nil
}

func (s *keyboard) ApplyBindings(ctx context.Context, changes []BindingChange) error {
	if err := s.acquire(ctx); err != nil {
		return err
	}
	defer s.release()
	changes = append([]BindingChange(nil), changes...)
	if err := ValidateBindings(changes); err != nil {
		return err
	}
	return s.applyBindings(ctx, changes)
}

// ValidateBindings checks a batch without performing device I/O.
func ValidateBindings(changes []BindingChange) error {
	seen := make(map[KeyID]bool, len(changes))
	for _, change := range changes {
		if err := validateChange(change); err != nil {
			return err
		}
		if seen[change.Key] {
			return fmt.Errorf("%w: duplicate key %s", ErrInvalidKey, change.Key)
		}
		seen[change.Key] = true
	}
	return nil
}

func (s *keyboard) applyBindings(ctx context.Context, changes []BindingChange) error {
	for i, change := range changes {
		if err := s.bind(ctx, change.Key, change.Action); err != nil {
			return &ApplyError{Completed: changes[:i:i], Failed: change, Err: err}
		}
	}
	return nil
}

// send is called only while holding the connection gate.
func (s *keyboard) send(ctx context.Context, cmd command.Command) ([][]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	for _, packet := range cmd.Build() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := s.transport.Write(ctx, packet); err != nil {
			return nil, fmt.Errorf("%s write failed: %w", cmd.Name(), err)
		}
	}
	response, err := cmd.ReadStrategy()(ctx, s.transport.Read)
	if err != nil {
		if errors.Is(err, command.ErrMalformedResponse) {
			return nil, fmt.Errorf("%w: %w", ErrMalformedResponse, err)
		}
		return nil, fmt.Errorf("%s read failed: %w", cmd.Name(), err)
	}
	return response, nil
}

func (s *keyboard) sendResponse(ctx context.Context, cmd command.Command, response command.Response) error {
	packets, err := s.send(ctx, cmd)
	if err != nil {
		return err
	}
	response.SetPayload(packets)
	return nil
}

func (s *keyboard) Close() error {
	<-s.gate
	defer s.release()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.transport.Close()
}
