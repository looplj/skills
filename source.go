package skills

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	githubTreeRe = regexp.MustCompile(`^/([^/]+)/([^/]+)/tree/([^/]+)(/.*)?$`)
	gitlabTreeRe = regexp.MustCompile(`^/([^/]+)/(.+)/-/tree/([^/]+)(/.*)?$`)
	ownerRepoRe  = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
)

func ParseSource(input string) (SkillSource, error) {
	in := strings.TrimSpace(input)
	if in == "" {
		return SkillSource{}, ErrSourceRequired
	}

	if looksLikeLocalPath(in) {
		p := expandHome(in)
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return SkillSource{Type: SkillSourceTypeLocal, SourceURL: abs}, nil
		}
	}

	if strings.HasPrefix(in, "http://") || strings.HasPrefix(in, "https://") {
		u, err := url.Parse(in)
		if err != nil {
			return SkillSource{}, err
		}

		host := strings.ToLower(u.Hostname())

		if host == "github.com" {
			m := githubTreeRe.FindStringSubmatch(u.Path)
			if len(m) > 0 {
				sub := strings.TrimPrefix(m[4], "/")

				return SkillSource{
					Type:      SkillSourceTypeGitHub,
					SourceURL: in,
					Owner:     m[1],
					Repo:      strings.TrimSuffix(m[2], ".git"),
					Ref:       m[3],
					Subpath:   sub,
				}, nil
			}

			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) >= 2 {
				return SkillSource{
					Type:      SkillSourceTypeGitHub,
					SourceURL: in,
					Owner:     parts[0],
					Repo:      strings.TrimSuffix(parts[1], ".git"),
				}, nil
			}
		}

		if host == "gitlab.com" || strings.Contains(host, "gitlab") {
			m := gitlabTreeRe.FindStringSubmatch(u.Path)
			if len(m) > 0 {
				sub := strings.TrimPrefix(m[4], "/")
				repoPath := strings.TrimSuffix(m[2], ".git")

				return SkillSource{
					Type:      SkillSourceTypeGitLab,
					SourceURL: in,
					Owner:     m[1],
					Repo:      repoPath,
					Ref:       m[3],
					Subpath:   sub,
				}, nil
			}
		}

		if strings.HasSuffix(strings.ToLower(u.Path), "/skill.md") {
			return SkillSource{Type: SkillSourceTypeDirectURL, SourceURL: in}, nil
		}

		if strings.HasSuffix(strings.ToLower(u.Path), ".git") {
			return SkillSource{Type: SkillSourceTypeGit, SourceURL: in}, nil
		}

		if host != "huggingface.co" && host != "github.com" && host != "gitlab.com" {
			return SkillSource{Type: SkillSourceTypeWellKnown, SourceURL: in, WellKnownURL: in}, nil
		}

		return SkillSource{Type: SkillSourceTypeGit, SourceURL: in}, nil
	}

	repoPart := in
	skillFilter := ""

	if strings.Count(in, "@") == 1 && strings.Index(in, "@") > 0 {
		left, right, _ := strings.Cut(in, "@")
		repoPart, skillFilter = left, right
	}

	if ownerRepoRe.MatchString(repoPart) {
		owner, repo, _ := strings.Cut(repoPart, "/")

		return SkillSource{
			Type:        SkillSourceTypeGitHub,
			SourceURL:   "https://github.com/" + repoPart,
			Owner:       owner,
			Repo:        repo,
			SkillFilter: skillFilter,
		}, nil
	}

	return SkillSource{}, ErrUnrecognizedSource
}

func looksLikeLocalPath(p string) bool {
	if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") || strings.HasPrefix(p, "/") {
		return true
	}

	if strings.HasPrefix(p, "~") {
		return true
	}

	if filepath.IsAbs(p) {
		return true
	}

	if _, err := os.Stat(p); err == nil {
		return true
	}

	return false
}

func expandHome(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}

	home, err := HomeDir()
	if err != nil {
		return p
	}

	if p == "~" {
		return home
	}

	if after, ok := strings.CutPrefix(p, "~/"); ok {
		return filepath.Join(home, after)
	}

	return p
}
