package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSource_LocalDir(t *testing.T) {
	tmp := t.TempDir()

	src, err := ParseSource(tmp)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if src.Type != SkillSourceTypeLocal {
		t.Fatalf("expected local, got %v", src.Type)
	}
}

func TestParseSource_GitHubShorthand(t *testing.T) {
	src, err := ParseSource("vercel-labs/agent-skills")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if src.Type != SkillSourceTypeGitHub {
		t.Fatalf("expected github, got %v", src.Type)
	}

	if src.Owner != "vercel-labs" || src.Repo != "agent-skills" {
		t.Fatalf("owner/repo mismatch: %s/%s", src.Owner, src.Repo)
	}

	if src.Ref != "" {
		t.Fatalf("expected empty ref for default branch discovery, got %q", src.Ref)
	}
}

func TestParseSource_GitHubRepoURL(t *testing.T) {
	src, err := ParseSource("https://github.com/a/b")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if src.Type != SkillSourceTypeGitHub || src.Ref != "" {
		t.Fatalf("expected github source with empty ref, got %#v", src)
	}
}

func TestParseSource_GitHubTreeURL(t *testing.T) {
	src, err := ParseSource("https://github.com/a/b/tree/main/skills/x")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if src.Type != SkillSourceTypeGitHub || src.Ref != "main" || src.Subpath != "skills/x" {
		t.Fatalf("unexpected parsed: %#v", src)
	}
}

func TestParseSource_SkillMdURL(t *testing.T) {
	src, err := ParseSource("https://example.com/skill.md")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if src.Type != SkillSourceTypeDirectURL {
		t.Fatalf("expected direct-url, got %v", src.Type)
	}
}

func TestParseSource_TildeExpansionOnlyWhenExists(t *testing.T) {
	home, err := HomeDir()
	if err != nil {
		t.Skip("no home")
	}

	dir := filepath.Join(home, "skills-test-does-not-exist")
	_ = os.RemoveAll(dir)

	src, err := ParseSource("~/skills-test-does-not-exist")
	if err == nil {
		t.Fatalf("expected error, got %+v", src)
	}
}
