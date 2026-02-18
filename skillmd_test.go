package skills

import (
	"slices"
	"testing"
)

func TestParseSkillMarkdown_OK(t *testing.T) {
	content := `---
name: Example
description: Something
compatibility: Requires git, docker, jq, and access to the internet
allowed-tools: Bash(git:*) Bash(jq:*) Read
metadata:
  install-name: example-skill
---

# Hi
`

	s, err := ParseSkillMarkdown(content)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if s.Name != "Example" {
		t.Fatalf("name mismatch: %q", s.Name)
	}

	if s.Description != "Something" {
		t.Fatalf("description mismatch: %q", s.Description)
	}

	if s.Compatibility != "Requires git, docker, jq, and access to the internet" {
		t.Fatalf("compatibility mismatch: %q", s.Compatibility)
	}

	if !slices.Equal(s.AllowedTools, []string{"Bash(git:*)", "Bash(jq:*)", "Read"}) {
		t.Fatalf("allowed-tools mismatch: %#v", s.AllowedTools)
	}

	if s.Metadata == nil {
		t.Fatalf("expected metadata")
	}
}

func TestParseSkillMarkdown_AllowedTools_WhitespaceSplit(t *testing.T) {
	content := `---
name: Example
description: Something
allowed-tools:  Read   Bash(git:*)` + "\n  " + `Bash(jq:*)
---

# Hi
`

	s, err := ParseSkillMarkdown(content)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if !slices.Equal(s.AllowedTools, []string{"Read", "Bash(git:*)", "Bash(jq:*)"}) {
		t.Fatalf("allowed-tools mismatch: %#v", s.AllowedTools)
	}
}

func TestParseSkillMarkdown_MissingFrontmatter(t *testing.T) {
	_, err := ParseSkillMarkdown("# no frontmatter\n")
	if err == nil {
		t.Fatalf("expected error")
	}
}
