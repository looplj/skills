package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSkill(t *testing.T, root string, installName string, name string, description string) {
	t.Helper()

	dir := filepath.Join(root, installName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := "---\n" +
		"name: " + name + "\n" +
		"description: " + description + "\n" +
		"---\n"

	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestListDirsWorkspaceOverridesGlobal(t *testing.T) {
	globalDir := t.TempDir()
	workspaceDir := t.TempDir()

	writeSkill(t, globalDir, "seo-audit", "seo-audit", "from global")
	writeSkill(t, workspaceDir, "seo-audit", "seo-audit", "from workspace")

	items, err := List(ListOptions{Dirs: []string{globalDir, workspaceDir}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].InstallName != "seo-audit" {
		t.Fatalf("expected installName seo-audit, got %q", items[0].InstallName)
	}

	if items[0].Description != "from workspace" {
		t.Fatalf("expected workspace description, got %q", items[0].Description)
	}
}

func TestGetDirsWorkspaceOverridesGlobal(t *testing.T) {
	globalDir := t.TempDir()
	workspaceDir := t.TempDir()

	writeSkill(t, globalDir, "seo-audit", "seo-audit", "from global")
	writeSkill(t, workspaceDir, "seo-audit", "seo-audit", "from workspace")

	res, err := Get(GetOptions{Skill: "seo-audit", Dirs: []string{globalDir, workspaceDir}})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if res.Skill.Description != "from workspace" {
		t.Fatalf("expected workspace description, got %q", res.Skill.Description)
	}
}

func TestGetDirsBySkillNameWhenInstallNameDiffers(t *testing.T) {
	globalDir := t.TempDir()
	workspaceDir := t.TempDir()

	writeSkill(t, globalDir, "seo-audit-global", "seo-audit", "from global")
	writeSkill(t, workspaceDir, "seo-audit-workspace", "seo-audit", "from workspace")

	res, err := Get(GetOptions{Skill: "seo-audit", Dirs: []string{globalDir, workspaceDir}})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if res.InstallName != "seo-audit-workspace" {
		t.Fatalf("expected workspace installName, got %q", res.InstallName)
	}

	if res.Skill.Description != "from workspace" {
		t.Fatalf("expected workspace description, got %q", res.Skill.Description)
	}
}

func TestListDirsInstalledOverridesBundled(t *testing.T) {
	globalDir := t.TempDir()
	workspaceDir := t.TempDir()

	writeSkill(t, globalDir, "seo-audit", "seo-audit", "from global")
	writeSkill(t, workspaceDir, "seo-audit", "seo-audit", "from workspace")

	items, err := List(ListOptions{
		BundledSkills: []Skill{{Name: "seo-audit", Description: "from bundled"}},
		Dirs:          []string{globalDir, workspaceDir},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Description != "from workspace" {
		t.Fatalf("expected workspace description, got %q", items[0].Description)
	}
}

func TestGetDirsInstalledOverridesBundled(t *testing.T) {
	globalDir := t.TempDir()
	workspaceDir := t.TempDir()

	writeSkill(t, globalDir, "seo-audit", "seo-audit", "from global")
	writeSkill(t, workspaceDir, "seo-audit", "seo-audit", "from workspace")

	res, err := Get(GetOptions{
		Skill:         "seo-audit",
		BundledSkills: []Skill{{Name: "seo-audit", Description: "from bundled"}},
		Dirs:          []string{globalDir, workspaceDir},
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if res.Skill.Description != "from workspace" {
		t.Fatalf("expected workspace description, got %q", res.Skill.Description)
	}
}

func TestGetFallsBackToBundledSkill(t *testing.T) {
	res, err := Get(GetOptions{
		Skill:         "seo-audit",
		Dirs:          []string{t.TempDir()},
		BundledSkills: []Skill{{Name: "seo-audit", Description: "from bundled", Content: "bundled content"}},
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if res.Skill.Description != "from bundled" {
		t.Fatalf("expected bundled description, got %q", res.Skill.Description)
	}
}
