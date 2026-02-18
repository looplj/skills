package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

func resolveTargetDirs(dirs []string, cwd string, global bool, agents []AgentType, enableAgentDiscovery bool) ([]string, []AgentType, error) {
	if len(dirs) > 0 {
		return dedupeDirs(dirs), nil, nil
	}

	if !enableAgentDiscovery {
		return nil, nil, ErrAgentDiscoveryRequired
	}

	if len(agents) == 0 {
		var err error

		agents, err = DetectInstalledAgents(cwd)
		if err != nil {
			return nil, nil, err
		}
	}

	canonicalDir, err := CanonicalSkillsDir(global, cwd)
	if err != nil {
		return nil, nil, err
	}

	var out []string

	out = append(out, canonicalDir)

	for _, a := range agents {
		dir, err := AgentSkillsDir(a, global, cwd)
		if err != nil {
			return nil, nil, err
		}

		out = append(out, dir)
	}

	return dedupeDirs(out), agents, nil
}

func dedupeDirs(dirs []string) []string {
	seen := map[string]bool{}

	out := make([]string, 0, len(dirs))
	for _, d := range dirs {
		d = filepath.Clean(strings.TrimSpace(d))
		if d == "." || d == "" {
			continue
		}

		if seen[d] {
			continue
		}

		seen[d] = true
		out = append(out, d)
	}

	return out
}

// Add installs skills from opts.Source and returns what is available and/or installed.
//
// The source may be a local directory, a git repository reference, a direct URL, or a
// well-known skills index. Behavior is controlled by opts:
//   - If opts.ListOnly is true, the function only discovers and returns skills (no install).
//   - If opts.All is true, it implies non-interactive mode and installs all discovered skills.
//   - If opts.Mode is empty, it defaults to symlink-based installation.
//   - If opts.Dirs is provided, it installs into those directories directly and skips agent discovery.
//   - Otherwise, agent discovery must be enabled and target directories are derived from the
//     canonical skills dir plus per-agent skills dirs (agents may be auto-detected when omitted).
//
// When installing globally with agent discovery enabled, Add also records entries in the skill lock
// file so CheckUpdates/Update can detect and apply future updates for supported sources.
//
//nolint:maintidx
func Add(ctx context.Context, opts AddOptions) (AddResult, error) {
	if opts.All {
		opts.Yes = true
		opts.Skills = []string{"*"}
	}

	if opts.Mode == "" {
		opts.Mode = InstallModeSymlink
	}

	if opts.Source == "" {
		return AddResult{}, ErrSourceRequired
	}

	cwd, err := os.Getwd()
	if err != nil {
		return AddResult{}, err
	}

	src, err := ParseSource(opts.Source)
	if err != nil {
		return AddResult{}, err
	}

	switch src.Type {
	case SkillSourceTypeDirectURL:
		dir, skill, installName, err := FetchDirectSkillToDir(ctx, src.SourceURL)
		if err != nil {
			return AddResult{}, err
		}

		defer func() { _ = os.RemoveAll(dir) }()

		if opts.ListOnly {
			return AddResult{Available: []Skill{skill}}, nil
		}

		targetDirs, agents, err := resolveTargetDirs(opts.Dirs, cwd, opts.Global, opts.Agents, opts.EnableAgentDiscovery)
		if err != nil {
			return AddResult{}, err
		}

		if err := installFromDirToTargets(installName, dir, targetDirs, opts.Mode); err != nil {
			return AddResult{}, err
		}

		if opts.EnableAgentDiscovery && opts.Global {
			lock, err := ReadSkillLock()
			if err != nil {
				return AddResult{}, err
			}

			AddSkillToLock(lock, installName, LockEntry{
				Source:     src,
				SourceType: string(src.Type),
				SourceURL:  src.SourceURL,
			})

			if err := WriteSkillLock(lock); err != nil {
				return AddResult{}, err
			}
		}

		return AddResult{Installed: []InstalledSkill{{
			InstallName: installName,
			Name:        skill.Name,
			Description: skill.Description,
			Source:      src,
			Agents:      agents,
			Global:      opts.Global,
		}}}, nil

	case SkillSourceTypeWellKnown:
		idx, resolvedBase, err := FetchWellKnownIndex(ctx, src.WellKnownURL)
		if err != nil {
			return AddResult{}, err
		}

		var entries []WellKnownSkillEntry

		if len(opts.Skills) == 0 {
			if len(idx.Skills) == 1 {
				entries = []WellKnownSkillEntry{idx.Skills[0]}
			} else if opts.Yes {
				entries = idx.Skills
			} else {
				return AddResult{}, ErrMultipleWellKnown
			}
		} else if containsStar(opts.Skills) {
			entries = idx.Skills
		} else {
			want := toStringSet(opts.Skills)
			for _, e := range idx.Skills {
				if want[e.Name] {
					entries = append(entries, e)
				}
			}
		}

		targetDirs := []string(nil)
		agents := []AgentType(nil)

		if !opts.ListOnly {
			var err error

			targetDirs, agents, err = resolveTargetDirs(opts.Dirs, cwd, opts.Global, opts.Agents, opts.EnableAgentDiscovery)
			if err != nil {
				return AddResult{}, err
			}
		}

		var available []Skill

		for _, e := range entries {
			dir, s, installName, err := FetchWellKnownSkillToDir(ctx, resolvedBase, e)
			if err != nil {
				return AddResult{}, err
			}

			defer func() { _ = os.RemoveAll(dir) }()

			available = append(available, s)

			if opts.ListOnly {
				continue
			}

			if err := installFromDirToTargets(installName, dir, targetDirs, opts.Mode); err != nil {
				return AddResult{}, err
			}
		}

		if opts.ListOnly {
			return AddResult{Available: available}, nil
		}

		if opts.EnableAgentDiscovery && opts.Global {
			lock, err := ReadSkillLock()
			if err != nil {
				return AddResult{}, err
			}

			for _, e := range entries {
				AddSkillToLock(lock, e.Name, LockEntry{
					Source:     src,
					SourceType: string(src.Type),
					SourceURL:  src.SourceURL,
					SkillPath:  e.Name,
				})
			}

			if err := WriteSkillLock(lock); err != nil {
				return AddResult{}, err
			}
		}

		var installed []InstalledSkill
		for i, e := range entries {
			installed = append(installed, InstalledSkill{
				InstallName: e.Name,
				Name:        available[i].Name,
				Description: available[i].Description,
				Source:      src,
				Agents:      agents,
				Global:      opts.Global,
			})
		}

		return AddResult{Available: available, Installed: installed}, nil
	case SkillSourceTypeLocal, SkillSourceTypeGitHub, SkillSourceTypeGitLab, SkillSourceTypeGit:
		break
	case SkillSourceTypeSearchHint:
		return AddResult{}, ErrUnsupportedSourceType
	}

	var repo *clonedRepo

	baseDir := ""

	switch src.Type {
	case SkillSourceTypeLocal:
		baseDir = src.SourceURL
	case SkillSourceTypeGitHub, SkillSourceTypeGitLab, SkillSourceTypeGit:
		repo, err = cloneRepo(ctx, src)
		if err != nil {
			return AddResult{}, err
		}
		defer repo.Cleanup()

		baseDir, err = repo.ResolveSubdir(src.Subpath)
		if err != nil {
			return AddResult{}, err
		}
	case SkillSourceTypeDirectURL, SkillSourceTypeWellKnown:
		return AddResult{}, ErrUnsupportedSourceType
	case SkillSourceTypeSearchHint:
		return AddResult{}, ErrUnsupportedSourceType
	default:
		return AddResult{}, ErrUnsupportedSourceType
	}

	found, err := DiscoverSkills(baseDir, opts.FullDepth)
	if err != nil {
		return AddResult{}, err
	}

	if src.SkillFilter != "" {
		var filtered []Skill

		for _, s := range found {
			if strings.EqualFold(filepath.Base(s.Dir), src.SkillFilter) || strings.EqualFold(s.Name, src.SkillFilter) {
				filtered = append(filtered, s)
			}
		}

		found = filtered
	}

	if opts.ListOnly {
		return AddResult{Available: found}, nil
	}

	if len(found) == 0 {
		return AddResult{}, ErrNoSkillsFound
	}

	var selected []Skill

	if len(opts.Skills) == 0 {
		if len(found) == 1 || opts.Yes {
			selected = found
		} else {
			return AddResult{}, ErrMultipleSkills
		}
	} else if containsStar(opts.Skills) {
		selected = found
	} else {
		want := toStringSet(opts.Skills)

		for _, s := range found {
			in := filepath.Base(s.Dir)
			if want[in] || want[s.Name] {
				selected = append(selected, s)
			}
		}
	}

	var (
		installed []InstalledSkill
		lock      *SkillLock
	)

	if opts.EnableAgentDiscovery && opts.Global {
		lock, err = ReadSkillLock()
		if err != nil {
			return AddResult{}, err
		}
	}

	targetDirs, agents, err := resolveTargetDirs(opts.Dirs, cwd, opts.Global, opts.Agents, opts.EnableAgentDiscovery)
	if err != nil {
		return AddResult{}, err
	}

	for _, s := range selected {
		installName := filepath.Base(s.Dir)
		if err := installFromDirToTargets(installName, s.Dir, targetDirs, opts.Mode); err != nil {
			return AddResult{}, err
		}

		if lock != nil {
			entry := LockEntry{
				Source:     src,
				SourceType: string(src.Type),
				SourceURL:  src.SourceURL,
			}
			if repo != nil && src.Type == SkillSourceTypeGitHub && src.Owner != "" && src.Repo != "" {
				rel, err := filepath.Rel(repo.Dir, s.Dir)
				if err == nil {
					entry.SkillPath = filepath.ToSlash(rel)
					ownerRepo := src.Owner + "/" + src.Repo

					hash, err := FetchGitHubSkillFolderHash(ctx, ownerRepo, entry.SkillPath, GetGitHubToken())
					if err == nil {
						entry.SkillFolderHash = hash
					}
				}
			}

			AddSkillToLock(lock, installName, entry)
		}

		installed = append(installed, InstalledSkill{
			InstallName: installName,
			Name:        s.Name,
			Description: s.Description,
			Source:      src,
			Agents:      agents,
			Global:      opts.Global,
		})
	}

	if lock != nil {
		if err := WriteSkillLock(lock); err != nil {
			return AddResult{}, err
		}
	}

	return AddResult{Available: found, Installed: installed}, nil
}

// List returns installed skills.
//
// If opts.Dirs is provided, it lists install names from those directories directly. Otherwise,
// agent discovery must be enabled and List scans the canonical skills directory plus per-agent
// skills directories (filtered by opts.Agents when provided).
func List(opts ListOptions) ([]ListedSkill, error) {
	if len(opts.Dirs) > 0 {
		type item struct {
			skill Skill
		}

		items := map[string]*item{}

		for _, root := range dedupeDirs(opts.Dirs) {
			installNames, err := listSkillInstallNames(root)
			if err != nil {
				return nil, err
			}

			for _, in := range installNames {
				it := items[in]
				if it == nil {
					it = &item{}
					items[in] = it
				}

				s, err := readInstalledSkill(root, in)
				if err == nil {
					it.skill = s
				}
			}
		}

		var out []ListedSkill
		for installName, it := range items {
			out = append(out, ListedSkill{
				InstallName: installName,
				Name:        it.skill.Name,
				Description: it.skill.Description,
				Global:      opts.Global,
			})
		}

		sort.Slice(out, func(i, j int) bool {
			return out[i].InstallName < out[j].InstallName
		})

		return out, nil
	}

	if !opts.EnableAgentDiscovery {
		return nil, ErrAgentDiscoveryRequired
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	canonicalDir, err := CanonicalSkillsDir(opts.Global, cwd)
	if err != nil {
		return nil, err
	}

	agents := opts.Agents
	if len(agents) == 0 {
		agents, err = DetectInstalledAgents(cwd)
		if err != nil {
			return nil, err
		}
	}

	installNames, err := listSkillInstallNames(canonicalDir)
	if err != nil {
		return nil, err
	}

	var out []ListedSkill

	for _, in := range installNames {
		s, err := readInstalledSkill(canonicalDir, in)
		if err != nil {
			continue
		}

		var present []AgentType

		for _, a := range agents {
			dir, err := AgentSkillsDir(a, opts.Global, cwd)
			if err != nil {
				continue
			}

			if _, err := os.Lstat(filepath.Join(dir, in)); err == nil {
				present = append(present, a)
			}
		}

		out = append(out, ListedSkill{
			InstallName: in,
			Name:        s.Name,
			Description: s.Description,
			Global:      opts.Global,
			Agents:      present,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].InstallName < out[j].InstallName
	})

	return out, nil
}

// Get returns the full content and metadata of an installed skill by name.
//
// If opts.Dirs is provided, it searches only those directories. Otherwise, agent discovery must be
// enabled and Get searches the canonical skills directory first and then per-agent skills
// directories (filtered by opts.Agents when provided).
func Get(opts GetOptions) (GetResult, error) {
	if opts.Skill == "" {
		return GetResult{}, ErrSkillNameRequired
	}

	if len(opts.Dirs) > 0 {
		roots := dedupeDirs(opts.Dirs)
		for i := len(roots) - 1; i >= 0; i-- {
			root := roots[i]
			s, err := readInstalledSkill(root, opts.Skill)
			if err == nil {
				return GetResult{InstallName: opts.Skill, Skill: s}, nil
			}
		}

		for i := len(roots) - 1; i >= 0; i-- {
			root := roots[i]

			installNames, err := listSkillInstallNames(root)
			if err != nil {
				return GetResult{}, err
			}

			for _, in := range installNames {
				s, err := readInstalledSkill(root, in)
				if err != nil {
					continue
				}

				if s.Name == opts.Skill {
					return GetResult{InstallName: in, Skill: s}, nil
				}
			}
		}

		return GetResult{}, fmt.Errorf("%w: %s", ErrSkillNotFound, opts.Skill)
	}

	if !opts.EnableAgentDiscovery {
		return GetResult{}, ErrAgentDiscoveryRequired
	}

	cwd, err := os.Getwd()
	if err != nil {
		return GetResult{}, err
	}

	canonicalDir, err := CanonicalSkillsDir(opts.Global, cwd)
	if err != nil {
		return GetResult{}, err
	}

	s, err := readInstalledSkill(canonicalDir, opts.Skill)
	if err == nil {
		return GetResult{InstallName: opts.Skill, Skill: s}, nil
	}

	installNames, err := listSkillInstallNames(canonicalDir)
	if err != nil {
		return GetResult{}, err
	}

	for _, in := range installNames {
		s, err := readInstalledSkill(canonicalDir, in)
		if err != nil {
			continue
		}

		if s.Name == opts.Skill {
			return GetResult{InstallName: in, Skill: s}, nil
		}
	}

	agentList := opts.Agents
	if len(agentList) == 0 {
		agentList, err = DetectInstalledAgents(cwd)
		if err != nil {
			return GetResult{}, err
		}
	}

	for _, a := range agentList {
		dir, err := AgentSkillsDir(a, opts.Global, cwd)
		if err != nil {
			continue
		}

		s, err := readInstalledSkill(dir, opts.Skill)
		if err == nil {
			return GetResult{InstallName: opts.Skill, Skill: s}, nil
		}

		installNames, err := listSkillInstallNames(dir)
		if err != nil {
			return GetResult{}, err
		}

		for _, in := range installNames {
			s, err := readInstalledSkill(dir, in)
			if err != nil {
				continue
			}

			if s.Name == opts.Skill {
				return GetResult{InstallName: in, Skill: s}, nil
			}
		}
	}

	return GetResult{}, fmt.Errorf("%w: %s", ErrSkillNotFound, opts.Skill)
}

// Remove deletes installed skills and returns what was removed.
//
// If opts.Dirs is provided, it removes from those directories directly:
//   - opts.All removes all install names found under the given dirs.
//   - opts.Skills removes only the specified install names.
//
// Otherwise, agent discovery must be enabled and Remove deletes from the canonical skills
// directory and per-agent skills directories (filtered by opts.Agents when provided). When removing
// from global installs, it also updates the skill lock so removed entries no longer participate in
// update checks.
func Remove(opts RemoveOptions) (RemoveResult, error) {
	if len(opts.Dirs) > 0 {
		targetDirs := dedupeDirs(opts.Dirs)
		if len(targetDirs) == 0 {
			return RemoveResult{}, ErrNoTargetDirs
		}

		var targets []string

		if opts.All {
			set := map[string]bool{}

			for _, root := range targetDirs {
				installNames, err := listSkillInstallNames(root)
				if err != nil {
					return RemoveResult{}, err
				}

				for _, in := range installNames {
					set[in] = true
				}
			}

			for in := range set {
				targets = append(targets, in)
			}

			sort.Strings(targets)
		} else if len(opts.Skills) > 0 {
			targets = opts.Skills
		} else {
			return RemoveResult{}, ErrNoSkillsSpecified
		}

		var removed []InstalledSkill

		for _, installName := range targets {
			for _, root := range targetDirs {
				_ = removeIfExists(filepath.Join(root, installName))
			}

			removed = append(removed, InstalledSkill{InstallName: installName})
		}

		return RemoveResult{Removed: removed}, nil
	}

	if !opts.EnableAgentDiscovery {
		return RemoveResult{}, ErrAgentDiscoveryRequired
	}

	cwd, err := os.Getwd()
	if err != nil {
		return RemoveResult{}, err
	}

	canonicalDir, err := CanonicalSkillsDir(opts.Global, cwd)
	if err != nil {
		return RemoveResult{}, err
	}

	agents := opts.Agents
	if len(agents) == 0 {
		agents, err = DetectInstalledAgents(cwd)
		if err != nil {
			return RemoveResult{}, err
		}
	}

	var targets []string
	if opts.All {
		targets, err = listSkillInstallNames(canonicalDir)
		if err != nil {
			return RemoveResult{}, err
		}
	} else if len(opts.Skills) > 0 {
		targets = opts.Skills
	} else {
		return RemoveResult{}, ErrNoSkillsSpecified
	}

	lock, err := ReadSkillLock()
	if err != nil {
		return RemoveResult{}, err
	}

	var removed []InstalledSkill

	for _, installName := range targets {
		for _, a := range agents {
			dir, err := AgentSkillsDir(a, opts.Global, cwd)
			if err != nil {
				continue
			}

			_ = removeIfExists(filepath.Join(dir, installName))
		}

		_ = removeIfExists(filepath.Join(canonicalDir, installName))

		if opts.Global {
			RemoveSkillFromLock(lock, installName)
		}

		removed = append(removed, InstalledSkill{InstallName: installName, Global: opts.Global, Agents: agents})
	}

	if opts.Global {
		if err := WriteSkillLock(lock); err != nil {
			return RemoveResult{}, err
		}
	}

	return RemoveResult{Removed: removed}, nil
}

// Init creates a SKILL.md template on disk and returns the created file path.
//
// By default it writes "SKILL.md" in the working directory. If opts.Name is provided, it writes
// "<dir>/<name>/SKILL.md". If opts.Dir is provided, it is used as the base directory instead of
// the current working directory.
func Init(opts InitOptions) (string, error) {
	dir := opts.Dir
	if dir == "" {
		var err error

		dir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	target := filepath.Join(dir, "SKILL.md")
	if opts.Name != "" {
		target = filepath.Join(dir, opts.Name, "SKILL.md")
		if err := ensureDir(filepath.Dir(target)); err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(target); err == nil {
		return "", ErrSkillAlreadyExists
	}

	name := opts.Name
	if name == "" {
		name = "My Skill"
	}

	content := strings.Join([]string{
		"---",
		"name: " + name,
		"description: A brief description of what this skill does",
		"---",
		"",
		"# Overview",
		"",
		"Describe the skill and when to use it.",
		"",
		"# Instructions",
		"",
		"Provide step-by-step guidance the agent should follow.",
		"",
	}, "\n")

	if err := ensureDir(filepath.Dir(target)); err != nil {
		return "", err
	}

	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		return "", err
	}

	return target, nil
}

// Search searches the skills registry and returns matching results.
func Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return SearchSkills(ctx, query, limit)
}

// CheckUpdates compares globally installed skills recorded in the skill lock against their remote
// source state and returns any detected updates.
//
// Currently, this only checks GitHub-based installs that have both a recorded skill folder path
// and a recorded folder hash.
func CheckUpdates(ctx context.Context) ([]UpdateCheck, error) {
	lock, err := ReadSkillLock()
	if err != nil {
		return nil, err
	}

	if len(lock.Skills) == 0 {
		return nil, nil
	}

	var out []UpdateCheck

	token := GetGitHubToken()

	for installName, entry := range lock.Skills {
		if entry.SourceType != string(SkillSourceTypeGitHub) {
			continue
		}

		if entry.SkillPath == "" || entry.SkillFolderHash == "" {
			continue
		}

		ownerRepo := entry.Source.Owner + "/" + entry.Source.Repo
		if ownerRepo == "/" || ownerRepo == "" {
			continue
		}

		remote, err := FetchGitHubSkillFolderHash(ctx, ownerRepo, entry.SkillPath, token)
		if err != nil {
			continue
		}

		if remote != entry.SkillFolderHash {
			out = append(out, UpdateCheck{
				InstallName: installName,
				Entry:       entry,
				CurrentHash: entry.SkillFolderHash,
				RemoteHash:  remote,
			})
		}
	}

	return out, nil
}

// Update applies updates for skills reported by CheckUpdates and returns the post-update status.
//
// Updates are applied by re-installing the recorded skill path from GitHub into the global target
// directories, using symlink install mode and non-interactive defaults.
func Update(ctx context.Context) ([]UpdateCheck, error) {
	updates, err := CheckUpdates(ctx)
	if err != nil {
		return nil, err
	}

	for _, u := range updates {
		owner := u.Entry.Source.Owner

		repo := u.Entry.Source.Repo
		if owner == "" || repo == "" {
			continue
		}

		ref := u.Entry.Source.Ref
		if ref == "" {
			ref = "main"
		}

		srcURL := "https://github.com/" + owner + "/" + repo + "/tree/" + ref + "/" + strings.TrimPrefix(u.Entry.SkillPath, "/")
		_, _ = Add(ctx, AddOptions{
			Source:               srcURL,
			Global:               true,
			Yes:                  true,
			Skills:               []string{u.InstallName},
			Mode:                 InstallModeSymlink,
			FullDepth:            true,
			EnableAgentDiscovery: true,
		})
	}

	return CheckUpdates(ctx)
}

func containsStar(list []string) bool {
	return slices.Contains(list, "*")
}

func toStringSet(list []string) map[string]bool {
	out := make(map[string]bool, len(list))
	for _, v := range list {
		out[strings.TrimSpace(v)] = true
	}

	return out
}
