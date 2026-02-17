package skills

import (
	"errors"
	"strings"

	"go.yaml.in/yaml/v3"
)

type skillFrontmatter struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Metadata    map[string]any `yaml:"metadata"`
}

func ParseSkillMarkdown(content string) (Skill, error) {
	fm, body, ok := splitFrontmatter(content)
	if !ok {
		return Skill{}, errors.New("missing frontmatter")
	}

	var parsed skillFrontmatter
	if err := yaml.Unmarshal([]byte(fm), &parsed); err != nil {
		return Skill{}, err
	}

	if strings.TrimSpace(parsed.Name) == "" {
		return Skill{}, errors.New("frontmatter missing name")
	}

	if strings.TrimSpace(parsed.Description) == "" {
		return Skill{}, errors.New("frontmatter missing description")
	}

	_ = body

	return Skill{
		Name:        parsed.Name,
		Description: parsed.Description,
		Content:     content,
		Metadata:    parsed.Metadata,
	}, nil
}

func splitFrontmatter(content string) (frontmatter string, rest string, ok bool) {
	s := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return "", content, false
	}

	after := strings.TrimPrefix(s, "---\n")

	before, after0, ok0 := strings.Cut(after, "\n---\n")
	if !ok0 {
		return "", content, false
	}

	fm := before
	rest = after0

	return fm, rest, true
}
