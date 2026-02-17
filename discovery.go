package skills

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func DiscoverSkills(baseDir string, fullDepth bool) ([]Skill, error) {
	rootSkill := filepath.Join(baseDir, "SKILL.md")
	if !fullDepth {
		if b, err := os.ReadFile(rootSkill); err == nil {
			s, err := ParseSkillMarkdown(string(b))
			if err != nil {
				return nil, err
			}

			s.Dir = baseDir

			return []Skill{s}, nil
		}
	}

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, nil
	}

	var out []Skill

	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.EqualFold(d.Name(), "SKILL.md") {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		s, err := ParseSkillMarkdown(string(b))
		if err != nil {
			return err
		}

		s.Dir = filepath.Dir(path)
		out = append(out, s)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
