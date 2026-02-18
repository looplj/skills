package skills

import (
	"encoding/json"
	"time"
)

type InstallMode string

const (
	InstallModeSymlink InstallMode = "symlink"
	InstallModeCopy    InstallMode = "copy"
)

type AgentType string

type AgentConfig struct {
	Name              AgentType
	DisplayName       string
	ProjectSkillsDir  string
	GlobalSkillsDir   string
	ShowInUniversal   bool
	DetectProjectPath string
	DetectGlobalPath  string
}

type Skill struct {
	Name          string
	Description   string
	Compatibility string
	AllowedTools  []string
	Dir           string
	Content       string
	Metadata      map[string]any
}

type SkillSourceType string

const (
	SkillSourceTypeLocal      SkillSourceType = "local"
	SkillSourceTypeGitHub     SkillSourceType = "github"
	SkillSourceTypeGitLab     SkillSourceType = "gitlab"
	SkillSourceTypeGit        SkillSourceType = "git"
	SkillSourceTypeDirectURL  SkillSourceType = "direct-url"
	SkillSourceTypeWellKnown  SkillSourceType = "well-known"
	SkillSourceTypeSearchHint SkillSourceType = "search-hint"
)

type SkillSource struct {
	Type         SkillSourceType `json:"type"`
	SourceURL    string          `json:"source_url,omitempty"`
	Owner        string          `json:"owner,omitempty"`
	Repo         string          `json:"repo,omitempty"`
	Ref          string          `json:"ref,omitempty"`
	Subpath      string          `json:"subpath,omitempty"`
	SkillFilter  string          `json:"skill_filter,omitempty"`
	ProviderID   string          `json:"provider_id,omitempty"`
	InstallName  string          `json:"install_name,omitempty"`
	WellKnownURL string          `json:"well_known_url,omitempty"`
}

type SearchResult struct {
	ID       string `json:"id"`
	SkillID  string `json:"skill_id"`
	Name     string `json:"name"`
	Installs int    `json:"installs"`
	Source   string `json:"source"`
}

type AddOptions struct {
	// Source is the install source string (local dir, git repo reference, direct URL,
	// or a well-known index URL).
	Source string

	// Dirs are explicit target directories to install into. When provided, agent discovery is
	// bypassed and Add installs directly to these directories.
	Dirs []string

	// Global installs into the global skills locations rather than project-local ones.
	// When agent discovery is enabled, global installs may also be tracked in the lock file
	// for future update checks.
	Global bool
	// Agents is an optional allowlist of agent types to install for. When empty and agent
	// discovery is enabled, installed agents may be auto-detected.
	Agents []AgentType
	// Skills selects which skills to install from a multi-skill source. Entries may match
	// either the skill folder name or the parsed skill name. "*" selects all discovered skills.
	Skills []string
	// ListOnly discovers and returns available skills without installing anything.
	ListOnly bool
	// Yes enables non-interactive defaults when a choice is required (for example, when a
	// source contains multiple skills and no explicit selection was provided).
	Yes bool
	// All installs all discovered skills. In Add, this implies Yes and sets Skills to "*".
	All bool
	// FullDepth enables deep scanning for SKILL.md files within the source directory tree.
	FullDepth bool
	// Mode controls how skill files are installed (for example, via symlinks or copying).
	Mode InstallMode

	// EnableAgentDiscovery enables deriving target directories from the canonical skills dir
	// and agent-specific skills dirs when Dirs is not provided.
	EnableAgentDiscovery bool
}

type InstalledSkill struct {
	InstallName string
	Name        string
	Description string
	Source      SkillSource
	Agents      []AgentType
	Global      bool
}

type AddResult struct {
	Available []Skill
	Installed []InstalledSkill
}

type RemoveOptions struct {
	// Dirs are explicit target directories to remove from. When provided, agent discovery is
	// bypassed and Remove deletes directly from these directories.
	Dirs []string
	// Global removes from the global install locations rather than project-local ones.
	Global bool
	// Agents is an optional allowlist of agent types to remove from. When empty and agent
	// discovery is enabled, installed agents may be auto-detected.
	Agents []AgentType
	// Skills is the list of install names to remove. When empty, All must be set.
	Skills []string
	// All removes all installed skills found in the resolved target directories.
	All bool
	// Yes is reserved for non-interactive defaults. It is currently unused by Remove.
	Yes bool

	// EnableAgentDiscovery enables deriving target directories from the canonical skills dir
	// and agent-specific skills dirs when Dirs is not provided.
	EnableAgentDiscovery bool
}

type RemoveResult struct {
	Removed []InstalledSkill
}

type ListOptions struct {
	// Dirs are explicit directories to list install names from. When provided, agent discovery
	// is bypassed and List reads only these directories.
	Dirs []string

	// Global lists from the global install locations rather than project-local ones.
	Global bool
	// Agents is an optional allowlist of agent types to list for. When empty and agent
	// discovery is enabled, installed agents may be auto-detected.
	Agents []AgentType

	// EnableAgentDiscovery enables deriving target directories from the canonical skills dir
	// and agent-specific skills dirs when Dirs is not provided.
	EnableAgentDiscovery bool
}

type ListedSkill struct {
	InstallName string
	Name        string
	Description string
	Global      bool
	Agents      []AgentType
}

type GetOptions struct {
	// Dirs are explicit directories to search. When provided, agent discovery is bypassed and
	// Get searches only these directories.
	Dirs []string

	// Skill is the install name of the skill to load.
	Skill string
	// Global searches the global install locations rather than project-local ones.
	Global bool
	// Agents is an optional allowlist of agent types to search. When empty and agent discovery
	// is enabled, installed agents may be auto-detected.
	Agents []AgentType

	// EnableAgentDiscovery enables deriving target directories from the canonical skills dir
	// and agent-specific skills dirs when Dirs is not provided.
	EnableAgentDiscovery bool
}

type GetResult struct {
	InstallName string
	Skill       Skill
}

type InitOptions struct {
	// Name is an optional folder name used to place the template under "<dir>/<name>/SKILL.md".
	Name string
	// Dir is the base directory to write into. When empty, the current working directory is used.
	Dir string
}

type LockEntry struct {
	Source          SkillSource `json:"source"`
	SourceType      string      `json:"source_type"`
	SourceURL       string      `json:"source_url"`
	SkillPath       string      `json:"skill_path,omitempty"`
	SkillFolderHash string      `json:"skill_folder_hash,omitempty"`
	InstalledAt     time.Time   `json:"installed_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type SkillLock struct {
	Skills             map[string]LockEntry `json:"skills"`
	LastSelectedAgents []AgentType          `json:"last_selected_agents,omitempty"`
	Dismissed          map[string]bool      `json:"dismissed,omitempty"`
}

func (s *SkillSource) UnmarshalJSON(b []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	type snake struct {
		Type         SkillSourceType `json:"type"`
		SourceURL    string          `json:"source_url,omitempty"`
		Owner        string          `json:"owner,omitempty"`
		Repo         string          `json:"repo,omitempty"`
		Ref          string          `json:"ref,omitempty"`
		Subpath      string          `json:"subpath,omitempty"`
		SkillFilter  string          `json:"skill_filter,omitempty"`
		ProviderID   string          `json:"provider_id,omitempty"`
		InstallName  string          `json:"install_name,omitempty"`
		WellKnownURL string          `json:"well_known_url,omitempty"`
	}

	type camel struct {
		Type         SkillSourceType `json:"type"`
		SourceURL    string          `json:"sourceUrl,omitempty"`
		Owner        string          `json:"owner,omitempty"`
		Repo         string          `json:"repo,omitempty"`
		Ref          string          `json:"ref,omitempty"`
		Subpath      string          `json:"subpath,omitempty"`
		SkillFilter  string          `json:"skillFilter,omitempty"`
		ProviderID   string          `json:"providerId,omitempty"`
		InstallName  string          `json:"installName,omitempty"`
		WellKnownURL string          `json:"wellKnownUrl,omitempty"`
	}

	if _, ok := obj["source_url"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["skill_filter"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["provider_id"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["install_name"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["well_known_url"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["sourceUrl"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["skillFilter"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["providerId"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["installName"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	if _, ok := obj["wellKnownUrl"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*s = SkillSource(v)

		return nil
	}

	var v snake
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	*s = SkillSource(v)

	return nil
}

func (r *SearchResult) UnmarshalJSON(b []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	type snake struct {
		ID       string `json:"id"`
		SkillID  string `json:"skill_id"`
		Name     string `json:"name"`
		Installs int    `json:"installs"`
		Source   string `json:"source"`
	}

	type camel struct {
		ID       string `json:"id"`
		SkillID  string `json:"skillId"`
		Name     string `json:"name"`
		Installs int    `json:"installs"`
		Source   string `json:"source"`
	}

	if _, ok := obj["skill_id"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*r = SearchResult(v)

		return nil
	}

	if _, ok := obj["skillId"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*r = SearchResult(v)

		return nil
	}

	var v snake
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	*r = SearchResult(v)

	return nil
}

func (e *LockEntry) UnmarshalJSON(b []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	type snake struct {
		Source          SkillSource `json:"source"`
		SourceType      string      `json:"source_type"`
		SourceURL       string      `json:"source_url"`
		SkillPath       string      `json:"skill_path,omitempty"`
		SkillFolderHash string      `json:"skill_folder_hash,omitempty"`
		InstalledAt     time.Time   `json:"installed_at"`
		UpdatedAt       time.Time   `json:"updated_at"`
	}

	type camel struct {
		Source          SkillSource `json:"source"`
		SourceType      string      `json:"sourceType"`
		SourceURL       string      `json:"sourceUrl"`
		SkillPath       string      `json:"skillPath,omitempty"`
		SkillFolderHash string      `json:"skillFolderHash,omitempty"`
		InstalledAt     time.Time   `json:"installedAt"`
		UpdatedAt       time.Time   `json:"updatedAt"`
	}

	if _, ok := obj["source_type"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*e = LockEntry(v)

		return nil
	}

	if _, ok := obj["installed_at"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*e = LockEntry(v)

		return nil
	}

	if _, ok := obj["sourceType"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*e = LockEntry(v)

		return nil
	}

	if _, ok := obj["installedAt"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*e = LockEntry(v)

		return nil
	}

	var v snake
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	*e = LockEntry(v)

	return nil
}

func (l *SkillLock) UnmarshalJSON(b []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	type snake struct {
		Skills             map[string]LockEntry `json:"skills"`
		LastSelectedAgents []AgentType          `json:"last_selected_agents,omitempty"`
		Dismissed          map[string]bool      `json:"dismissed,omitempty"`
	}

	type camel struct {
		Skills             map[string]LockEntry `json:"skills"`
		LastSelectedAgents []AgentType          `json:"lastSelectedAgents,omitempty"`
		Dismissed          map[string]bool      `json:"dismissed,omitempty"`
	}

	if _, ok := obj["last_selected_agents"]; ok {
		var v snake
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*l = SkillLock(v)

		return nil
	}

	if _, ok := obj["lastSelectedAgents"]; ok {
		var v camel
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}

		*l = SkillLock(v)

		return nil
	}

	var v snake
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	*l = SkillLock(v)

	return nil
}

type UpdateCheck struct {
	InstallName string
	Entry       LockEntry
	CurrentHash string
	RemoteHash  string
}
