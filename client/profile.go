package client

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/ColinEge/8bd-ultimate-api/internal/command"
)

// ValidateProfileName bounds names to the two observed 29-byte text packets.
func ValidateProfileName(title string) error {
	if !utf8.ValidString(title) || strings.TrimSpace(title) == "" ||
		strings.ContainsRune(title, 0) || len(utf16.Encode([]rune(title))) > 29 {
		return fmt.Errorf("%w: title must be nonblank, contain no NUL, and fit in 29 UTF-16 code units", ErrInvalidProfileName)
	}
	return nil
}

func (s *keyboard) GetProfile(ctx context.Context) (string, error) {
	if err := s.acquire(ctx); err != nil {
		return "", err
	}
	defer s.release()
	var response command.ProfileResponse
	if err := s.sendResponse(ctx, &command.GetProfile{}, &response); err != nil {
		return "", fmt.Errorf("%w: %w", ErrProfileQueryFailed, err)
	}
	title, err := response.Decode()
	if err != nil {
		return "", fmt.Errorf("%w: %w: %w", ErrProfileQueryFailed, ErrMalformedResponse, err)
	}
	return title, nil
}

func (s *keyboard) SetProfile(ctx context.Context, title string) error {
	if err := s.acquire(ctx); err != nil {
		return err
	}
	defer s.release()
	if err := ValidateProfileName(title); err != nil {
		return err
	}
	return s.setProfile(ctx, title)
}

func (s *keyboard) setProfile(ctx context.Context, title string) error {
	var response command.StatusResponse
	if err := s.sendResponse(ctx, &command.SetProfile{ProfileName: title}, &response); err != nil {
		return fmt.Errorf("%w: %w", ErrProfileMutationFailed, err)
	}
	status, err := response.DecodeStatus()
	if err != nil {
		return fmt.Errorf("%w: %w: %w", ErrProfileMutationFailed, ErrMalformedResponse, err)
	}
	if status != command.StatusResponseOk {
		return fmt.Errorf("%w: %w: set profile responded %s", ErrProfileMutationFailed, ErrDeviceRejected, status)
	}
	return nil
}
