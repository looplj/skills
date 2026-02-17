package skills

import (
	"os"
	"path/filepath"
)

func HomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if home == "" {
		return "", ErrHomeNotFound
	}

	return home, nil
}

func XDGConfigHome() (string, error) {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v, nil
	}

	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config"), nil
}

func CodexHome() (string, error) {
	if v := os.Getenv("CODEX_HOME"); v != "" {
		return v, nil
	}

	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".codex"), nil
}

func ClaudeHome() (string, error) {
	if v := os.Getenv("CLAUDE_CONFIG_DIR"); v != "" {
		return v, nil
	}

	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".claude"), nil
}

func CanonicalSkillsDir(global bool, cwd string) (string, error) {
	if global {
		home, err := HomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, ".agents", "skills"), nil
	}

	if cwd == "" {
		var err error

		cwd, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(cwd, ".agents", "skills"), nil
}

func SkillLockPath() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".agents", ".skill-lock.json"), nil
}
