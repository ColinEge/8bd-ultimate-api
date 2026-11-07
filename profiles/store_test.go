package profiles

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/client"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(WithDirectory(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestStoreCRUD(t *testing.T) {
	s := testStore(t)
	profile := Profile{Title: "Gaming", Keys: []client.BindingChange{
		{Key: client.KeyA, Action: client.ActionB()}, {Key: client.KeyC, Action: client.ActionDisabled()},
	}}
	if err := s.Create(profile); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(profile); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("duplicate: %v", err)
	}
	profile.Keys[0].Action = client.ActionC()
	if err := s.Save(profile); err != nil {
		t.Fatal(err)
	}
	reopened, err := NewStore(WithDirectory(s.Directory()))
	if err != nil {
		t.Fatal(err)
	}
	got, err := reopened.Load("Gaming")
	if err != nil || !reflect.DeepEqual(got, profile) {
		t.Fatalf("persisted=%+v, error=%v", got, err)
	}
	if err := s.Create(Profile{Title: "Work"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Rename("Gaming", "Work"); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("rename collision: %v", err)
	}
	if err := s.Rename("Gaming", "FPS"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load("Gaming"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("old name: %v", err)
	}
	got, err = s.Load("FPS")
	profile.Title = "FPS"
	if err != nil || !reflect.DeepEqual(got, profile) {
		t.Fatalf("renamed=%+v, error=%v", got, err)
	}
	list, err := s.List()
	if err != nil || len(list) != 2 || list[0].Title != "FPS" || list[1].Title != "Work" {
		t.Fatalf("list=%+v, error=%v", list, err)
	}
	list[0].Keys[0].Action = client.ActionA()
	got, err = s.Load("FPS")
	if err != nil || got.Keys[0].Action != client.ActionC() {
		t.Fatalf("list shared state: %+v, %v", got, err)
	}
	if err := s.Delete("FPS"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("FPS"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing delete: %v", err)
	}
	if err := s.Save(Profile{Title: "Missing"}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("save created missing profile: %v", err)
	}
	entries, err := os.ReadDir(s.Directory())
	if err != nil || len(entries) != 1 || strings.HasPrefix(entries[0].Name(), ".profile-") {
		t.Fatalf("temporary file leak: %v, %v", entries, err)
	}
}

func TestStorePaths(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux default path")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	path, err := DefaultDirectory()
	if err != nil || path != filepath.Join(home, ".local", "share", "8bd-kbd-api", "profiles") {
		t.Fatalf("default=%q, error=%v", path, err)
	}
	xdg := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdg)
	s, err := NewStore()
	if err != nil || s.Directory() != filepath.Join(xdg, "8bd-kbd-api", "profiles") {
		t.Fatalf("XDG=%v, error=%v", s, err)
	}
	t.Setenv("XDG_DATA_HOME", "relative")
	if _, err := NewStore(); err == nil {
		t.Fatal("relative XDG accepted")
	}
	override := filepath.Join(t.TempDir(), "custom")
	s, err = NewStore(WithDirectory(override))
	if err != nil || s.Directory() != override {
		t.Fatalf("override=%v, error=%v", s, err)
	}
	if _, err := NewStore(WithDirectory("")); err == nil {
		t.Fatal("empty override accepted")
	}
	if _, err := NewStore(nil); err == nil {
		t.Fatal("nil option accepted")
	}
}

func TestStoreValidation(t *testing.T) {
	s := testStore(t)
	for _, data := range []string{
		`{`, `null`, `{"version":2,"title":"Gaming","keys":[]}`,
		`{"version":1,"title":"Gaming","keys":null}`,
		`{"version":1,"title":"Gaming","keys":[],"unknown":true}`,
		`{"version":1,"title":"Gaming","keys":[]} {}`,
		`{"version":1,"title":"Other","keys":[]}`,
		`{"version":1,"title":"Gaming","keys":[{"key":4,"action":"broken"}]}`,
		`{"version":1,"title":"Gaming","keys":[{"key":0,"action":"07-00-05"}]}`,
		`{"version":1,"title":"Gaming","keys":[{"key":4,"action":"07-00-ff"}]}`,
		`{"version":1,"title":"Gaming","keys":[{"key":4,"action":"07-00-05"},{"key":4,"action":"07-00-06"}]}`,
	} {
		if err := os.WriteFile(s.path("Gaming"), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Load("Gaming"); !errors.Is(err, ErrInvalidProfile) {
			t.Fatalf("data=%s, error=%v", data, err)
		}
		if _, err := s.List(); !errors.Is(err, ErrInvalidProfile) {
			t.Fatalf("corrupt profile skipped: %v", err)
		}
	}
	if err := s.Delete("Gaming"); err != nil {
		t.Fatal(err)
	}
	for _, profile := range []Profile{
		{Title: ""}, {Title: strings.Repeat("a", 30)},
		{Title: "Gaming", Keys: []client.BindingChange{{Key: client.KeyA}}},
	} {
		if err := s.Create(profile); !errors.Is(err, ErrInvalidProfile) {
			t.Fatalf("invalid profile accepted: %+v, %v", profile, err)
		}
	}
	profile := Profile{Title: "../Work"}
	if err := s.Create(profile); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Load(profile.Title); err != nil || got.Title != profile.Title {
		t.Fatalf("path-like title: %+v, %v", got, err)
	}
	info, err := os.Stat(s.path(profile.Title))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("permissions: %v, %v", info, err)
	}
}

func TestConcurrentCreate(t *testing.T) {
	s := testStore(t)
	other, err := NewStore(WithDirectory(s.Directory()))
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, store := range []*Store{s, other} {
		wg.Go(func() { results <- store.Create(Profile{Title: "Gaming"}) })
	}
	wg.Wait()
	close(results)
	succeeded, duplicates := 0, 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, fs.ErrExist):
			duplicates++
		default:
			t.Fatal(err)
		}
	}
	if succeeded != 1 || duplicates != 1 {
		t.Fatalf("success=%d, duplicate=%d", succeeded, duplicates)
	}
}

func TestStoreIOErrors(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewStore(WithDirectory(file)); err == nil {
		t.Fatal("file accepted as directory")
	}
	s := testStore(t)
	if err := os.Remove(s.Directory()); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(Profile{Title: "Gaming"}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing directory create: %v", err)
	}
	if _, err := s.List(); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing directory list: %v", err)
	}
}
