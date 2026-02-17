package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func SkillsAPIURL() string {
	if v := strings.TrimSpace(os.Getenv("SKILLS_API_URL")); v != "" {
		return v
	}

	return "https://skills.sh"
}

func SearchSkills(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}

	if limit <= 0 {
		limit = 10
	}

	base, err := url.Parse(SkillsAPIURL())
	if err != nil {
		return nil, err
	}

	base.Path = strings.TrimSuffix(base.Path, "/") + "/api/search"
	v := base.Query()
	v.Set("q", q)
	v.Set("limit", strconv.Itoa(limit))
	base.RawQuery = v.Encode()

	body, err := httpGetText(ctx, base.String())
	if err != nil {
		return nil, err
	}

	return parseSearchResultsJSON([]byte(body))
}

func parseSearchResultsJSON(b []byte) ([]SearchResult, error) {
	type searchResponse struct {
		Skills []SearchResult `json:"skills"`
	}

	var resp searchResponse

	errObject := json.Unmarshal(b, &resp)
	if errObject == nil && resp.Skills != nil {
		return resp.Skills, nil
	}

	var results []SearchResult

	errArray := json.Unmarshal(b, &results)
	if errArray == nil {
		return results, nil
	}

	if errObject == nil {
		return nil, fmt.Errorf("unexpected search response shape")
	}

	return nil, errObject
}
