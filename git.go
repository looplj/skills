package skills

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type clonedRepo struct {
	Dir string
}

func cloneRepo(ctx context.Context, src SkillSource) (*clonedRepo, error) {
	tmp, err := os.MkdirTemp("", "skills-repo-*")
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		_ = os.RemoveAll(tmp)
	}

	repoURL := src.SourceURL
	if src.Type == SkillSourceTypeGitHub && src.Owner != "" && src.Repo != "" && !strings.HasPrefix(repoURL, "http") {
		repoURL = "https://github.com/" + src.Owner + "/" + src.Repo + ".git"
	}

	if src.Type == SkillSourceTypeGitHub && src.Owner != "" && src.Repo != "" && strings.HasPrefix(repoURL, "https://github.com/") && !strings.HasSuffix(repoURL, ".git") {
		repoURL = "https://github.com/" + src.Owner + "/" + src.Repo + ".git"
	}

	if src.Type == SkillSourceTypeGitLab && strings.HasPrefix(repoURL, "https://gitlab.com/") && !strings.HasSuffix(repoURL, ".git") {
		repoURL = repoURL + ".git"
	}

	ref := src.Ref
	if ref == "" {
		ref = "main"
	}

	args := []string{"clone", "--depth", "1", "--branch", ref, repoURL, tmp}
	cmd := exec.CommandContext(ctx, "git", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return nil, errors.New("git clone failed: " + strings.TrimSpace(string(out)))
	}

	return &clonedRepo{Dir: tmp}, nil
}

func (r *clonedRepo) ResolveSubdir(subpath string) (string, error) {
	if subpath == "" {
		return r.Dir, nil
	}

	p := filepath.Join(r.Dir, filepath.FromSlash(subpath))
	if _, err := os.Stat(p); err != nil {
		return "", err
	}

	return p, nil
}

func (r *clonedRepo) Cleanup() {
	_ = os.RemoveAll(r.Dir)
}
