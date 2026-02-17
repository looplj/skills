package skills

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type gitHubTreeResponse struct {
	SHA  string `json:"sha"`
	Tree []struct {
		Path string `json:"path"`
		Type string `json:"type"`
		SHA  string `json:"sha"`
	} `json:"tree"`
}

func FetchGitHubSkillFolderHash(ctx context.Context, ownerRepo string, skillPath string, token string) (string, error) {
	folderPath := strings.ReplaceAll(skillPath, "\\", "/")
	if before, ok := strings.CutSuffix(folderPath, "/SKILL.md"); ok {
		folderPath = before
	} else if before, ok := strings.CutSuffix(folderPath, "SKILL.md"); ok {
		folderPath = before
	}

	folderPath = strings.TrimSuffix(folderPath, "/")

	branches := []string{"main", "master"}

	var lastErr error

	for _, branch := range branches {
		u := "https://api.github.com/repos/" + ownerRepo + "/git/trees/" + branch + "?recursive=1"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return "", err
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "skills-go")

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := defaultHTTPClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = errors.New(resp.Status)
			continue
		}

		var data gitHubTreeResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			lastErr = err
			continue
		}

		if folderPath == "" {
			if data.SHA == "" {
				lastErr = errors.New("missing tree sha")
				continue
			}

			return data.SHA, nil
		}

		for _, e := range data.Tree {
			if e.Type == "tree" && e.Path == folderPath {
				if e.SHA == "" {
					break
				}

				return e.SHA, nil
			}
		}

		lastErr = errors.New("folder not found in tree")
	}

	if lastErr == nil {
		lastErr = errors.New("failed to fetch github tree")
	}

	return "", lastErr
}
