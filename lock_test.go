package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSkillLock_ReadWriteAndUpdate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	lock, err := ReadSkillLock()
	if err != nil {
		t.Fatal(err)
	}

	if lock.Skills == nil {
		t.Fatalf("expected skills map initialized")
	}

	if len(lock.Skills) != 0 {
		t.Fatalf("expected empty lock, got %d", len(lock.Skills))
	}

	AddSkillToLock(lock, "a-skill", LockEntry{
		SourceType: "local",
		SourceURL:  "/tmp/src",
	})

	if err := WriteSkillLock(lock); err != nil {
		t.Fatal(err)
	}

	lockPath, err := SkillLockPath()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("expected lock file written: %v", err)
	}

	lock2, err := ReadSkillLock()
	if err != nil {
		t.Fatal(err)
	}

	e := lock2.Skills["a-skill"]
	if e.SourceType != "local" || e.SourceURL != "/tmp/src" {
		t.Fatalf("unexpected entry: %+v", e)
	}

	if e.InstalledAt.IsZero() || e.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps set: %+v", e)
	}

	oldInstalledAt := e.InstalledAt
	oldUpdatedAt := e.UpdatedAt

	time.Sleep(5 * time.Millisecond)

	AddSkillToLock(lock2, "a-skill", LockEntry{
		SourceType: "local",
		SourceURL:  "/tmp/changed",
	})

	e2 := lock2.Skills["a-skill"]
	if !e2.InstalledAt.Equal(oldInstalledAt) {
		t.Fatalf("expected InstalledAt preserved; old=%v new=%v", oldInstalledAt, e2.InstalledAt)
	}

	if e2.SourceURL != "/tmp/changed" {
		t.Fatalf("expected SourceURL updated, got %q", e2.SourceURL)
	}

	if !e2.UpdatedAt.After(oldUpdatedAt) {
		t.Fatalf("expected UpdatedAt to advance; old=%v new=%v", oldUpdatedAt, e2.UpdatedAt)
	}

	RemoveSkillFromLock(lock2, "a-skill")

	if _, ok := lock2.Skills["a-skill"]; ok {
		t.Fatalf("expected entry removed")
	}
}

func TestAdd_GlobalWritesSkillLock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	base := t.TempDir()
	writeSkillMarkdown(t, filepath.Join(base, "alpha"), "Alpha", "A")
	writeSkillMarkdown(t, filepath.Join(base, "beta"), "Beta", "B")

	target := t.TempDir()

	_, err := Add(context.Background(), AddOptions{
		Source:               base,
		Yes:                  true,
		Global:               true,
		EnableAgentDiscovery: true,
		Dirs:                 []string{target},
		Mode:                 InstallModeCopy,
	})
	if err != nil {
		t.Fatal(err)
	}

	lock, err := ReadSkillLock()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := lock.Skills["alpha"]; !ok {
		t.Fatalf("expected alpha in lock")
	}

	if _, ok := lock.Skills["beta"]; !ok {
		t.Fatalf("expected beta in lock")
	}
}
