package profiles

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/ColinEge/8bd-kbd-api/client"
)

var ErrInvalidProfile = errors.New("invalid local profile")

type Store struct {
	mu        sync.Mutex
	directory string
}

type StoreOption func(*Store) error

// WithDirectory overrides the complete profile directory, not its parent.
func WithDirectory(directory string) StoreOption {
	return func(s *Store) error {
		if strings.TrimSpace(directory) == "" {
			return fmt.Errorf("profile directory must not be empty")
		}
		path, err := filepath.Abs(directory)
		if err != nil {
			return err
		}
		s.directory = path
		return nil
	}
}

func DefaultDirectory() (string, error) {
	var base string
	var err error
	switch runtime.GOOS {
	case "windows":
		base, err = os.UserCacheDir()
	case "darwin":
		base, err = os.UserHomeDir()
		base = filepath.Join(base, "Library", "Application Support")
	default:
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			base, err = os.UserHomeDir()
			base = filepath.Join(base, ".local", "share")
		}
	}
	if err != nil {
		return "", fmt.Errorf("resolve user data directory: %w", err)
	}
	if !filepath.IsAbs(base) {
		return "", fmt.Errorf("user data directory must be absolute: %q", base)
	}
	return filepath.Join(base, "8bd-kbd-api", "profiles"), nil
}

func NewStore(options ...StoreOption) (*Store, error) {
	s := &Store{}
	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("nil profile store option")
		}
		if err := option(s); err != nil {
			return nil, err
		}
	}
	if s.directory == "" {
		path, err := DefaultDirectory()
		if err != nil {
			return nil, err
		}
		s.directory = path
	}
	if err := os.MkdirAll(s.directory, 0700); err != nil {
		return nil, fmt.Errorf("create profile directory: %w", err)
	}
	return s, nil
}

func (s *Store) Directory() string { return s.directory }

// Filenames are derived from titles so a title cannot escape the directory.
func (s *Store) path(title string) string {
	return filepath.Join(s.directory, fmt.Sprintf("%x.json", sha256.Sum256([]byte(title))))
}

type diskKey struct {
	Key    uint8  `json:"key"`
	Action string `json:"action"`
}

type diskProfile struct {
	Version int       `json:"version"`
	Title   string    `json:"title"`
	Keys    []diskKey `json:"keys"`
}

func encode(profile Profile) ([]byte, error) {
	if err := Validate(profile); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidProfile, err)
	}
	disk := diskProfile{Version: 1, Title: profile.Title, Keys: make([]diskKey, 0, len(profile.Keys))}
	for _, key := range profile.Keys {
		disk.Keys = append(disk.Keys, diskKey{Key: uint8(key.Key), Action: key.Action.String()})
	}
	return json.MarshalIndent(disk, "", "  ")
}

func decode(data []byte) (Profile, error) {
	var disk diskProfile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&disk); err != nil {
		return Profile{}, fmt.Errorf("%w: %w", ErrInvalidProfile, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Profile{}, fmt.Errorf("%w: expected one JSON object", ErrInvalidProfile)
	}
	if disk.Version != 1 || disk.Keys == nil {
		return Profile{}, fmt.Errorf("%w: expected version 1 and a keys array", ErrInvalidProfile)
	}
	profile := Profile{Title: disk.Title, Keys: make([]client.BindingChange, 0, len(disk.Keys))}
	for _, key := range disk.Keys {
		action, err := client.ParseAction(key.Action)
		if err != nil {
			return Profile{}, fmt.Errorf("%w: %w", ErrInvalidProfile, err)
		}
		profile.Keys = append(profile.Keys, client.BindingChange{Key: client.KeyID(key.Key), Action: action})
	}
	if err := Validate(profile); err != nil {
		return Profile{}, fmt.Errorf("%w: %w", ErrInvalidProfile, err)
	}
	return profile, nil
}

func (s *Store) Load(title string) (Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load(title)
}

func (s *Store) load(title string) (Profile, error) {
	if err := client.ValidateProfileName(title); err != nil {
		return Profile{}, err
	}
	data, err := os.ReadFile(s.path(title))
	if err != nil {
		return Profile{}, fmt.Errorf("load profile %q: %w", title, err)
	}
	profile, err := decode(data)
	if err != nil {
		return Profile{}, fmt.Errorf("load profile %q: %w", title, err)
	}
	if profile.Title != title {
		return Profile{}, fmt.Errorf("%w: stored title does not match %q", ErrInvalidProfile, title)
	}
	return profile, nil
}

func (s *Store) List() ([]Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return nil, err
	}
	profiles := make([]Profile, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		profile, err := decode(data)
		if err != nil {
			return nil, fmt.Errorf("list profile %s: %w", entry.Name(), err)
		}
		if filepath.Base(s.path(profile.Title)) != entry.Name() {
			return nil, fmt.Errorf("%w: filename does not match title in %s", ErrInvalidProfile, entry.Name())
		}
		profiles = append(profiles, profile)
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Title < profiles[j].Title })
	return profiles, nil
}

func (s *Store) Create(profile Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.write(profile, false)
}

// Save updates an existing profile; it does not implicitly create one.
func (s *Store) Save(profile Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.load(profile.Title); err != nil {
		return err
	}
	return s.write(profile, true)
}

func (s *Store) write(profile Profile, replace bool) (err error) {
	data, err := encode(profile)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(s.directory, ".profile-")
	if err != nil {
		return err
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
			err = errors.Join(err, removeErr)
		}
	}()
	_, writeErr := file.Write(data)
	if writeErr == nil {
		writeErr = file.Sync()
	}
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	if replace {
		err = os.Rename(file.Name(), s.path(profile.Title))
	} else {
		// Link publishes a complete file without overwriting a duplicate title.
		err = os.Link(file.Name(), s.path(profile.Title))
	}
	if err != nil {
		return fmt.Errorf("save profile %q: %w", profile.Title, err)
	}
	return nil
}

func (s *Store) Rename(title, newTitle string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	profile, err := s.load(title)
	if err != nil {
		return err
	}
	if title == newTitle {
		return nil
	}
	profile.Title = newTitle
	if err := s.write(profile, false); err != nil {
		return err
	}
	if err := os.Remove(s.path(title)); err != nil {
		return errors.Join(fmt.Errorf("remove old profile after rename: %w", err), os.Remove(s.path(newTitle)))
	}
	return nil
}

func (s *Store) Delete(title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := client.ValidateProfileName(title); err != nil {
		return err
	}
	if err := os.Remove(s.path(title)); err != nil {
		return fmt.Errorf("delete profile %q: %w", title, err)
	}
	return nil
}
