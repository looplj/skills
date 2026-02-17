package skills

import "testing"

func TestParseSearchResultsJSON_Object(t *testing.T) {
	b := []byte(`{"query":"marketing","skills":[{"id":"o/r/s","skillId":"s","name":"seo-audit","installs":123,"source":"o/r"}]}`)

	got, err := parseSearchResultsJSON(b)
	if err != nil {
		t.Fatalf("parseSearchResultsJSON() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(got))
	}

	if got[0].ID != "o/r/s" {
		t.Fatalf("results[0].ID = %q, want %q", got[0].ID, "o/r/s")
	}

	if got[0].SkillID != "s" {
		t.Fatalf("results[0].SkillID = %q, want %q", got[0].SkillID, "s")
	}

	if got[0].Name != "seo-audit" {
		t.Fatalf("results[0].Name = %q, want %q", got[0].Name, "seo-audit")
	}

	if got[0].Installs != 123 {
		t.Fatalf("results[0].Installs = %d, want %d", got[0].Installs, 123)
	}

	if got[0].Source != "o/r" {
		t.Fatalf("results[0].Source = %q, want %q", got[0].Source, "o/r")
	}
}

func TestParseSearchResultsJSON_Array(t *testing.T) {
	b := []byte(`[{"id":"o/r/s","skillId":"s","name":"seo-audit","installs":123,"source":"o/r"}]`)

	got, err := parseSearchResultsJSON(b)
	if err != nil {
		t.Fatalf("parseSearchResultsJSON() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(got))
	}

	if got[0].ID != "o/r/s" {
		t.Fatalf("results[0].ID = %q, want %q", got[0].ID, "o/r/s")
	}
}

func TestParseSearchResultsJSON_ObjectMissingSkills(t *testing.T) {
	_, err := parseSearchResultsJSON([]byte(`{"count":0}`))
	if err == nil {
		t.Fatalf("parseSearchResultsJSON() error = nil, want non-nil")
	}
}
