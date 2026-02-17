package skills

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type WellKnownIndex struct {
	Skills []WellKnownSkillEntry `json:"skills"`
}

type WellKnownSkillEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
}

func FetchDirectSkillToDir(ctx context.Context, skillURL string) (dir string, skill Skill, installName string, err error) {
	content, err := httpGetText(ctx, skillURL)
	if err != nil {
		return "", Skill{}, "", err
	}

	s, err := ParseSkillMarkdown(content)
	if err != nil {
		return "", Skill{}, "", err
	}

	tmp, err := os.MkdirTemp("", "skills-direct-*")
	if err != nil {
		return "", Skill{}, "", err
	}

	if err := os.WriteFile(filepath.Join(tmp, "SKILL.md"), []byte(content), 0o644); err != nil {
		_ = os.RemoveAll(tmp)
		return "", Skill{}, "", err
	}

	in := strings.TrimSpace(s.Name)
	if v, ok := s.Metadata["install-name"]; ok {
		if vv, ok := v.(string); ok && strings.TrimSpace(vv) != "" {
			in = strings.TrimSpace(vv)
		}
	}

	s.Dir = tmp

	return tmp, s, in, nil
}

func FetchWellKnownIndex(ctx context.Context, baseURL string) (index WellKnownIndex, resolvedBase string, err error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return WellKnownIndex{}, "", err
	}

	basePath := strings.TrimSuffix(u.Path, "/")

	try := []string{
		u.Scheme + "://" + u.Host + basePath + "/.well-known/skills/index.json",
	}
	if basePath != "" {
		try = append(try, u.Scheme+"://"+u.Host+"/.well-known/skills/index.json")
	}

	for _, idxURL := range try {
		var idx WellKnownIndex
		if err := httpGetJSON(ctx, idxURL, &idx); err != nil {
			continue
		}

		if len(idx.Skills) == 0 {
			continue
		}

		valid := true

		for _, s := range idx.Skills {
			if strings.TrimSpace(s.Name) == "" || strings.TrimSpace(s.Description) == "" || len(s.Files) == 0 {
				valid = false
				break
			}

			has := false

			for _, f := range s.Files {
				if !isSafeRelPath(f) {
					valid = false
					break
				}

				if strings.EqualFold(f, "SKILL.md") {
					has = true
				}
			}

			if !valid || !has {
				valid = false
				break
			}
		}

		if !valid {
			continue
		}

		if strings.Contains(idxURL, basePath+"/.well-known/skills/index.json") {
			return idx, u.Scheme + "://" + u.Host + basePath, nil
		}

		return idx, u.Scheme + "://" + u.Host, nil
	}

	return WellKnownIndex{}, "", errors.New("no well-known skills index found")
}

func FetchWellKnownSkillToDir(ctx context.Context, resolvedBase string, entry WellKnownSkillEntry) (dir string, skill Skill, installName string, err error) {
	skillBase := strings.TrimSuffix(resolvedBase, "/") + "/.well-known/skills/" + entry.Name

	tmp, err := os.MkdirTemp("", "skills-wellknown-*")
	if err != nil {
		return "", Skill{}, "", err
	}

	cleanup := func() {
		_ = os.RemoveAll(tmp)
	}

	for _, f := range entry.Files {
		if !isSafeRelPath(f) {
			cleanup()
			return "", Skill{}, "", errors.New("unsafe file path in well-known index")
		}

		content, err := httpGetText(ctx, skillBase+"/"+f)
		if err != nil {
			if strings.EqualFold(f, "SKILL.md") {
				cleanup()
				return "", Skill{}, "", err
			}

			continue
		}

		dst := filepath.Join(tmp, filepath.FromSlash(f))
		if err := ensureDir(filepath.Dir(dst)); err != nil {
			cleanup()
			return "", Skill{}, "", err
		}

		if err := os.WriteFile(dst, []byte(content), 0o644); err != nil {
			cleanup()
			return "", Skill{}, "", err
		}
	}

	b, err := os.ReadFile(filepath.Join(tmp, "SKILL.md"))
	if err != nil {
		cleanup()
		return "", Skill{}, "", err
	}

	s, err := ParseSkillMarkdown(string(b))
	if err != nil {
		cleanup()
		return "", Skill{}, "", err
	}

	s.Dir = tmp

	return tmp, s, entry.Name, nil
}

func isSafeRelPath(p string) bool {
	if p == "" {
		return false
	}

	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "\\") {
		return false
	}

	if strings.Contains(p, "..") {
		return false
	}

	return true
}
