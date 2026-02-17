package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSkills_RootOnly(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "SKILL.md"), []byte(`---
name: Root
description: D
---
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(tmp, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmp, "nested", "SKILL.md"), []byte(`---
name: Nested
description: D
---
`), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := DiscoverSkills(tmp, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(skills) != 1 || skills[0].Name != "Root" {
		t.Fatalf("unexpected: %#v", skills)
	}
}

func TestDiscoverSkills_FullDepth(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmp, "nested", "SKILL.md"), []byte(`---
name: Nested
description: D
---
`), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := DiscoverSkills(tmp, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(skills) != 1 || skills[0].Name != "Nested" {
		t.Fatalf("unexpected: %#v", skills)
	}
}
