package skills

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectDefaultBranch(t *testing.T) {
	tmp := t.TempDir()
	remote := filepath.Join(tmp, "remote.git")
	work := filepath.Join(tmp, "work")

	runCmd(t, tmp, "git", "init", "--bare", remote)
	runCmd(t, tmp, "git", "clone", remote, work)
	runCmd(t, work, "git", "checkout", "-b", "master")
	runCmd(t, work, "git", "config", "user.name", "tester")
	runCmd(t, work, "git", "config", "user.email", "tester@example.com")
	runCmd(t, work, "git", "commit", "--allow-empty", "-m", "init")
	runCmd(t, work, "git", "push", "-u", "origin", "master")

	branch, err := detectDefaultBranch(context.Background(), remote)
	if err != nil {
		t.Fatalf("detectDefaultBranch returned error: %v", err)
	}

	if branch != "master" {
		t.Fatalf("expected master, got %q", branch)
	}
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, out)
	}
}
