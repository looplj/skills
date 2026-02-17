package skills

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func ReadSkillLock() (*SkillLock, error) {
	path, err := SkillLockPath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &SkillLock{Skills: map[string]LockEntry{}}, nil
		}

		return nil, err
	}

	var lock SkillLock
	if err := json.Unmarshal(b, &lock); err != nil {
		return nil, err
	}

	if lock.Skills == nil {
		lock.Skills = map[string]LockEntry{}
	}

	return &lock, nil
}

func WriteSkillLock(lock *SkillLock) error {
	path, err := SkillLockPath()
	if err != nil {
		return err
	}

	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}

	if lock.Skills == nil {
		lock.Skills = map[string]LockEntry{}
	}

	b, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}

	b = append(b, '\n')

	return os.WriteFile(path, b, 0o644)
}

func AddSkillToLock(lock *SkillLock, installName string, entry LockEntry) {
	if lock.Skills == nil {
		lock.Skills = map[string]LockEntry{}
	}

	now := time.Now()

	if existing, ok := lock.Skills[installName]; ok {
		entry.InstalledAt = existing.InstalledAt
	} else {
		entry.InstalledAt = now
	}

	entry.UpdatedAt = now
	lock.Skills[installName] = entry
}

func RemoveSkillFromLock(lock *SkillLock, installName string) {
	if lock.Skills == nil {
		return
	}

	delete(lock.Skills, installName)
}

func GetGitHubToken() string {
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		return v
	}

	if v := os.Getenv("GH_TOKEN"); v != "" {
		return v
	}

	out, err := exec.CommandContext(context.Background(), "gh", "auth", "token").CombinedOutput()
	if err == nil {
		if t := strings.TrimSpace(string(out)); t != "" {
			return t
		}
	}

	return ""
}
