package skills

import "testing"

func TestParseSkillMarkdown_OK(t *testing.T) {
	content := `---
name: Example
description: Something
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

	if s.Metadata == nil {
		t.Fatalf("expected metadata")
	}
}

func TestParseSkillMarkdown_MissingFrontmatter(t *testing.T) {
	_, err := ParseSkillMarkdown("# no frontmatter\n")
	if err == nil {
		t.Fatalf("expected error")
	}
}
