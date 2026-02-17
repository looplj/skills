package skills

import (
	"os"
	"path/filepath"
	"sort"
)

func listSkillInstallNames(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	var out []string

	for _, e := range ents {
		if !e.IsDir() {
			continue
		}

		out = append(out, e.Name())
	}

	sort.Strings(out)

	return out, nil
}

func readInstalledSkill(dir string, installName string) (Skill, error) {
	path := filepath.Join(dir, installName, "SKILL.md")

	b, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}

	s, err := ParseSkillMarkdown(string(b))
	if err != nil {
		return Skill{}, err
	}

	s.Dir = filepath.Join(dir, installName)

	return s, nil
}
