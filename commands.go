package skills

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

func resolveTargetDirs(cwd string, global bool, dirs []string, agents []AgentType, enableAgentDiscovery bool) ([]string, []AgentType, error) {
	if len(dirs) > 0 {
		return dedupeDirs(dirs), nil, nil
	}

	if !enableAgentDiscovery {
		return nil, nil, errors.New("target dirs are required when agent discovery is disabled")
	}

	agentList := agents
	if len(agentList) == 0 {
		var err error

		agentList, err = DetectInstalledAgents(cwd)
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

	for _, a := range agentList {
		dir, err := AgentSkillsDir(a, global, cwd)
		if err != nil {
			return nil, nil, err
		}

		out = append(out, dir)
	}

	return dedupeDirs(out), agentList, nil
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
		return AddResult{}, errors.New("source is required")
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

		targetDirs, agents, err := resolveTargetDirs(cwd, opts.Global, opts.Dirs, opts.Agents, opts.EnableAgentDiscovery)
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
				return AddResult{}, errors.New("multiple well-known skills found; specify --skill or --yes")
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

			targetDirs, agents, err = resolveTargetDirs(cwd, opts.Global, opts.Dirs, opts.Agents, opts.EnableAgentDiscovery)
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
		return AddResult{}, errors.New("unsupported source type: search-hint")
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
		return AddResult{}, errors.New("unsupported source type")
	case SkillSourceTypeSearchHint:
		return AddResult{}, errors.New("unsupported source type: search-hint")
	default:
		return AddResult{}, errors.New("unsupported source type")
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
		return AddResult{}, errors.New("no skills found")
	}

	var selected []Skill

	if len(opts.Skills) == 0 {
		if len(found) == 1 || opts.Yes {
			selected = found
		} else {
			return AddResult{}, errors.New("multiple skills found; specify --skill or --yes")
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

	targetDirs, agents, err := resolveTargetDirs(cwd, opts.Global, opts.Dirs, opts.Agents, opts.EnableAgentDiscovery)
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

func List(opts ListOptions) ([]ListedSkill, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if len(opts.Dirs) > 0 {
		type item struct {
			skill Skill
			dirs  []string
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

				it.dirs = append(it.dirs, root)
				if it.skill.Name == "" {
					s, err := readInstalledSkill(root, in)
					if err == nil {
						it.skill = s
					}
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
		return nil, errors.New("target dirs are required when agent discovery is disabled")
	}

	canonicalDir, err := CanonicalSkillsDir(opts.Global, cwd)
	if err != nil {
		return nil, err
	}

	agentList := opts.Agents
	if len(agentList) == 0 {
		agentList, err = DetectInstalledAgents(cwd)
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

		for _, a := range agentList {
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

func Remove(opts RemoveOptions) (RemoveResult, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return RemoveResult{}, err
	}

	if len(opts.Dirs) > 0 {
		targetDirs := dedupeDirs(opts.Dirs)
		if len(targetDirs) == 0 {
			return RemoveResult{}, errors.New("no target dirs specified")
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
			return RemoveResult{}, errors.New("no skills specified")
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
		return RemoveResult{}, errors.New("target dirs are required when agent discovery is disabled")
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
		return RemoveResult{}, errors.New("no skills specified")
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
		return "", errors.New("SKILL.md already exists")
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

func Find(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return SearchSkills(ctx, query, limit)
}

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
