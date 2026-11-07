// Package profiles stores local keyboard profiles independently of USB clients.
package profiles

import (
	"context"
	"fmt"

	"github.com/ColinEge/8bd-kbd-api/client"
)

// Profile contains a device title and only the bindings to change.
type Profile struct {
	Title string                 `json:"title"`
	Keys  []client.BindingChange `json:"keys"`
}

func Validate(profile Profile) error {
	if err := client.ValidateProfileName(profile.Title); err != nil {
		return err
	}
	return client.ValidateBindings(profile.Keys)
}

// ApplyError distinguishes an acknowledged title from a failed binding.
// The failing operation may have reached the device; read back before retrying.
type ApplyError struct {
	TitleSet bool
	Err      error
}

func (e *ApplyError) Error() string {
	return fmt.Sprintf("apply profile failed (title acknowledged: %t): %v", e.TitleSet, e.Err)
}

func (e *ApplyError) Unwrap() error { return e.Err }

// Apply validates before I/O, sets the title, then writes the saved keys in order.
// Omitted keys are untouched. There is no rollback or automatic retry.
// The caller owns the keyboard and must prevent concurrent edits during Apply.
func Apply(ctx context.Context, keyboard client.Keyboard, profile Profile) error {
	profile.Keys = append([]client.BindingChange(nil), profile.Keys...)
	if err := Validate(profile); err != nil {
		return err
	}
	if err := keyboard.SetProfile(ctx, profile.Title); err != nil {
		return &ApplyError{Err: err}
	}
	if err := keyboard.ApplyBindings(ctx, profile.Keys); err != nil {
		return &ApplyError{TitleSet: true, Err: err}
	}
	return nil
}

// Capture reads the device title and stored overrides without writing anything.
func Capture(ctx context.Context, keyboard client.Keyboard) (Profile, error) {
	title, err := keyboard.GetProfile(ctx)
	if err != nil {
		return Profile{}, err
	}
	bindings, err := keyboard.Bindings(ctx)
	if err != nil {
		return Profile{}, err
	}
	profile := Profile{Title: title, Keys: make([]client.BindingChange, 0, len(bindings))}
	for _, binding := range bindings {
		if binding.State != client.BindingDefault {
			profile.Keys = append(profile.Keys, client.BindingChange{Key: binding.Key, Action: binding.Action})
		}
	}
	if err := Validate(profile); err != nil {
		return Profile{}, fmt.Errorf("capture profile: %w", err)
	}
	return profile, nil
}
