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

	repoURL := normalizeRepoURL(src)
	ref, err := resolveCloneRef(ctx, repoURL, src.Ref)
	if err != nil {
		cleanup()
		return nil, err
	}

	args := []string{"clone", "--depth", "1"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, repoURL, tmp)

	cmd := exec.CommandContext(ctx, "git", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return nil, errors.New("git clone failed: " + strings.TrimSpace(string(out)))
	}

	return &clonedRepo{Dir: tmp}, nil
}

func normalizeRepoURL(src SkillSource) string {
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

	return repoURL
}

func resolveCloneRef(ctx context.Context, repoURL string, requestedRef string) (string, error) {
	if requestedRef != "" {
		return requestedRef, nil
	}

	ref, err := detectDefaultBranch(ctx, repoURL)
	if err == nil && ref != "" {
		return ref, nil
	}

	return "", nil
}

func detectDefaultBranch(ctx context.Context, repoURL string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--symref", repoURL, "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New("git ls-remote failed: " + strings.TrimSpace(string(out)))
	}

	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "ref:") || !strings.Contains(line, "\tHEAD") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		return strings.TrimPrefix(fields[1], "refs/heads/"), nil
	}

	return "", errors.New("default branch not found")
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
