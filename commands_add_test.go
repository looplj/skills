package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeSkillMarkdown(t *testing.T, dir string, name string, desc string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	b := []byte(`---
name: ` + name + `
description: ` + desc + `
---
`)
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAdd_ListOnly_LocalRootSkill(t *testing.T) {
	base := t.TempDir()
	writeSkillMarkdown(t, base, "Root", "Desc")

	res, err := Add(context.Background(), AddOptions{
		Source:   base,
		ListOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Installed) != 0 {
		t.Fatalf("expected no installs, got %d", len(res.Installed))
	}

	if len(res.Available) != 1 {
		t.Fatalf("expected 1 available, got %d", len(res.Available))
	}

	if res.Available[0].Name != "Root" {
		t.Fatalf("expected skill name Root, got %q", res.Available[0].Name)
	}
}

func TestAdd_NoSkillsFound(t *testing.T) {
	base := t.TempDir()

	_, err := Add(context.Background(), AddOptions{
		Source: base,
	})
	if err == nil {
		t.Fatalf("expected error")
	}

	if err.Error() != "no skills found" {
		t.Fatalf("expected no skills found error, got %v", err)
	}
}

func TestAdd_MultipleSkills_RequiresYesOrSkill(t *testing.T) {
	base := t.TempDir()
	writeSkillMarkdown(t, filepath.Join(base, "alpha"), "Alpha", "A")
	writeSkillMarkdown(t, filepath.Join(base, "beta"), "Beta", "B")

	_, err := Add(context.Background(), AddOptions{
		Source: base,
	})
	if err == nil {
		t.Fatalf("expected error")
	}

	if err.Error() != "multiple skills found; specify --skill or --yes" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdd_YesInstallsAllFound(t *testing.T) {
	base := t.TempDir()
	writeSkillMarkdown(t, filepath.Join(base, "alpha"), "Alpha", "A")
	writeSkillMarkdown(t, filepath.Join(base, "beta"), "Beta", "B")

	target := t.TempDir()

	res, err := Add(context.Background(), AddOptions{
		Source: base,
		Yes:    true,
		Dirs:   []string{target},
		Mode:   InstallModeCopy,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Installed) != 2 {
		t.Fatalf("expected 2 installs, got %d", len(res.Installed))
	}

	if _, err := os.Stat(filepath.Join(target, "alpha", "SKILL.md")); err != nil {
		t.Fatalf("expected alpha installed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "beta", "SKILL.md")); err != nil {
		t.Fatalf("expected beta installed: %v", err)
	}
}

func TestAdd_StarInstallsAllFound(t *testing.T) {
	base := t.TempDir()
	writeSkillMarkdown(t, filepath.Join(base, "alpha"), "Alpha", "A")
	writeSkillMarkdown(t, filepath.Join(base, "beta"), "Beta", "B")

	target := t.TempDir()

	res, err := Add(context.Background(), AddOptions{
		Source: base,
		Skills: []string{"*"},
		Dirs:   []string{target},
		Mode:   InstallModeCopy,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Installed) != 2 {
		t.Fatalf("expected 2 installs, got %d", len(res.Installed))
	}
}

func TestAdd_SelectByDirNameOrSkillName(t *testing.T) {
	base := t.TempDir()
	writeSkillMarkdown(t, filepath.Join(base, "alpha"), "Zed", "A")
	writeSkillMarkdown(t, filepath.Join(base, "beta"), "Other", "B")

	target := t.TempDir()

	res, err := Add(context.Background(), AddOptions{
		Source: base,
		Skills: []string{"Zed"},
		Dirs:   []string{target},
		Mode:   InstallModeCopy,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Installed) != 1 {
		t.Fatalf("expected 1 install, got %d", len(res.Installed))
	}

	if res.Installed[0].InstallName != "alpha" {
		t.Fatalf("expected alpha install name, got %q", res.Installed[0].InstallName)
	}

	if _, err := os.Stat(filepath.Join(target, "alpha", "SKILL.md")); err != nil {
		t.Fatalf("expected alpha installed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "beta", "SKILL.md")); err == nil {
		t.Fatalf("expected beta not installed")
	}
}
