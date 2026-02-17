package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallFromDir_CreatesCanonical(t *testing.T) {
	cwd := t.TempDir()

	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte(`---
name: A
description: B
---
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(src, "extra.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	canonical, err := CanonicalSkillsDir(false, cwd)
	if err != nil {
		t.Fatal(err)
	}

	if err := installFromDirToTargets("a-skill", src, []string{canonical}, InstallModeCopy); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(canonical, "a-skill", "SKILL.md")); err != nil {
		t.Fatalf("expected SKILL.md: %v", err)
	}

	if _, err := os.Stat(filepath.Join(canonical, "a-skill", "extra.txt")); err != nil {
		t.Fatalf("expected extra.txt: %v", err)
	}
}
