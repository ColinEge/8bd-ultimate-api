package client

import (
	"context"
	"fmt"

	"github.com/ColinEge/8bd-kbd-api/keydefs"
)

type KeyID = keydefs.KeyID
type Action = keydefs.ActionCode

type Modifier byte

const (
	LeftCtrl Modifier = 0xe0 + iota
	LeftShift
	LeftAlt
	LeftMeta
	RightCtrl
	RightShift
	RightAlt
	RightMeta
)

// Chord constructs the firmware's single-modifier keyboard action.
func Chord(modifier Modifier, action Action) (Action, error) {
	if modifier < LeftCtrl || modifier > RightMeta ||
		!action.IsKeyboardUsage() || action[1] != 0 || action[2] == 0 {
		return Action{}, fmt.Errorf("%w: chord requires one modifier and an ordinary keyboard action", ErrInvalidAction)
	}
	if !keydefs.SupportedAction(action) {
		return Action{}, fmt.Errorf("%w: chord action %s", ErrUnsupported, action)
	}
	action[1] = byte(modifier)
	return action, nil
}

// ParseAction parses raw firmware bytes; Bind additionally checks support.
func ParseAction(text string) (Action, error) {
	action, err := keydefs.ParseActionCode(text)
	if err != nil {
		return Action{}, fmt.Errorf("%w: %w", ErrInvalidAction, err)
	}
	return action, nil
}

type BindingState string

const (
	BindingDefault  BindingState = "default"
	BindingAction   BindingState = "action"
	BindingDisabled BindingState = "disabled"
)

type Binding struct {
	Key       KeyID
	State     BindingState
	Action    Action // Stored override; zero for BindingDefault.
	Effective Action
}

type BindingChange struct {
	Key    KeyID
	Action Action
}

// ApplyError counts acknowledged changes, not necessarily all device writes.
// Read back the failed key before retrying an operation with uncertain outcome.
type ApplyError struct {
	Completed []BindingChange
	Failed    BindingChange
	Err       error
}

func (e *ApplyError) Error() string {
	return fmt.Sprintf("apply bindings failed at key %s after %d acknowledged changes: %v", e.Failed.Key, len(e.Completed), e.Err)
}

func (e *ApplyError) Unwrap() error { return e.Err }

type KeyboardInfo struct {
	Device DeviceInfo
}

type KeyInfo struct {
	ID      KeyID
	Label   string
	Default Action
}

type ActionInfo struct {
	Action   Action
	Label    string
	Category keydefs.KeyCategory
}

// Keyboard exposes configuration operations without protocol commands.
// Operations on a connection are serialized; Close is idempotent.
type Keyboard interface {
	Info(ctx context.Context) (KeyboardInfo, error)
	FirmwareVersion(ctx context.Context) (string, error)

	Keys() []KeyInfo
	Actions() []ActionInfo

	GetProfile(ctx context.Context) (string, error)
	SetProfile(ctx context.Context, title string) error

	Bind(ctx context.Context, key KeyID, action Action) error
	Binding(ctx context.Context, key KeyID) (Binding, error)
	Bindings(ctx context.Context) ([]Binding, error)
	Disable(ctx context.Context, key KeyID) error
	Restore(ctx context.Context, key KeyID) error

	ApplyBindings(
		ctx context.Context,
		bindings []BindingChange,
	) error

	Close() error
}
