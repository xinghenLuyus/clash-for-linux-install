package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfilesFromMeta(t *testing.T) {
	home := t.TempDir()
	resources := filepath.Join(home, "resources")
	if err := os.MkdirAll(resources, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := `profiles:
  - id: 1
    path: /tmp/1.yaml
    url: https://example.com/a
  - id: "2"
    path: /tmp/2.yaml
    url: 'https://example.com/b'
use: 2
`
	if err := os.WriteFile(filepath.Join(resources, "profiles.yaml"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &app{home: home}
	profiles, err := a.loadProfilesFromMeta()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].ID != "1" || profiles[0].URL != "https://example.com/a" || profiles[0].Current {
		t.Fatalf("unexpected first profile: %#v", profiles[0])
	}
	if profiles[1].ID != "2" || profiles[1].URL != "https://example.com/b" || !profiles[1].Current {
		t.Fatalf("unexpected second profile: %#v", profiles[1])
	}
}
