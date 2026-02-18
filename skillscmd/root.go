package skillscmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	skills "github.com/looplj/skills"
)

type RootOptions struct {
	// WorkspaceDir is an explicit workspace skills directory to operate on (for example,
	// ".agents/skills"). When provided, commands can operate without agent discovery.
	WorkspaceDir string
	// GlobalDir is an explicit global skills directory to operate on (for example,
	// "~/.agents/skills"). When provided, commands can operate without agent discovery.
	GlobalDir string
	// Commands limits which subcommands are registered. When empty, all commands are registered.
	Commands []string
	// Use is the root command name shown in help/usage output.
	Use string
	// Stdout is where command output is written. Defaults to os.Stdout.
	Stdout *os.File
	// Stderr is where command errors/help are written. Defaults to os.Stderr.
	Stderr *os.File
	// EnableAgentDiscovery enables deriving target directories from known agent installs
	// when explicit Dirs are not provided.
	EnableAgentDiscovery bool
	// EnableAgentFlags enables registering agent-related flags (like --agent).
	// When unset, defaults to false.
	EnableAgentFlags bool
}

// NewRootCommand builds the CLI root command and registers all subcommands.
func NewRootCommand(opts RootOptions) *cobra.Command {
	rootOpts := opts
	if opts.Use == "" {
		rootOpts.Use = "skills"
	}
	stdout := rootOpts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := rootOpts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	root := &cobra.Command{
		Use:           rootOpts.Use,
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Manage AI agent skills",
	}

	enableAgentFlags := rootOpts.EnableAgentFlags
	rootUse := rootOpts.Use
	commandEnabled := func(name string) bool {
		if len(rootOpts.Commands) == 0 {
			return true
		}
		return slices.Contains(rootOpts.Commands, name)
	}

	if commandEnabled("search") {
		root.AddCommand(newSearchCmd(stdout, rootUse))
	}
	if commandEnabled("add") {
		root.AddCommand(newAddCmd(stdout, &rootOpts, enableAgentFlags, rootUse))
	}
	if commandEnabled("list") {
		root.AddCommand(newListCmd(stdout, &rootOpts, enableAgentFlags, rootUse))
	}
	if commandEnabled("get") {
		root.AddCommand(newGetCmd(stdout, &rootOpts, enableAgentFlags, rootUse))
	}
	if commandEnabled("remove") {
		root.AddCommand(newRemoveCmd(stdout, &rootOpts, enableAgentFlags, rootUse))
	}
	if commandEnabled("init") {
		root.AddCommand(newInitCmd(stdout, rootUse))
	}
	if commandEnabled("check") {
		root.AddCommand(newCheckCmd(stdout, rootUse))
	}
	if commandEnabled("update") {
		root.AddCommand(newUpdateCmd(rootUse))
	}

	root.SetOut(stdout)
	root.SetErr(stderr)

	root.PersistentFlags().StringVar(&rootOpts.WorkspaceDir, "workspace-dir", rootOpts.WorkspaceDir, "Workspace skills directory (e.g. .agents/skills)")
	root.PersistentFlags().StringVar(&rootOpts.GlobalDir, "global-dir", rootOpts.GlobalDir, "Global skills directory (e.g. ~/.agents/skills)")

	return root
}

// newSearchCmd creates the "search" command, which searches the skills registry.
func newSearchCmd(out *os.File, rootUse string) *cobra.Command {
	var limit int
	var format string
	var quiet bool
	cmd := &cobra.Command{
		Use:     "search <query>",
		Args:    cobra.MinimumNArgs(1),
		Short:   "Search skills from registry, search from https://skills.sh by default",
		Example: fmt.Sprintf("  %s search docker\n  %s search --limit 5 kubernetes\n  %s search docker --format add\n  %s add $(%s search docker --format add | head -n 1)", rootUse, rootUse, rootUse, rootUse, rootUse),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			results, err := skills.Search(ctx, strings.Join(args, " "), limit)
			if err != nil {
				return err
			}
			errOut := cmd.ErrOrStderr()
			switch format {
			case "", "table":
				fmt.Fprintln(out, "NAME\tID\tSKILL_ID\tSOURCE\tINSTALLS\tADD")
				for _, r := range results {
					addRef := fmt.Sprintf("%s@%s", r.Source, r.SkillID)
					fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%d\t%s\n", r.Name, r.ID, r.SkillID, r.Source, r.Installs, addRef)
				}
			case "add":
				if !quiet {
					fmt.Fprintln(errOut, "add 格式输出每行一个安装引用: <source>@<skill_id>")
					fmt.Fprintln(errOut, "示例:")
					fmt.Fprintf(errOut, "  %s search docker --format add | head -n 5\n", rootUse)
					fmt.Fprintf(errOut, "  %s add $(%s search docker --format add | head -n 1)\n", rootUse, rootUse)
					fmt.Fprintf(errOut, "  %s search docker --format add | head -n 5 | xargs -n 1 %s add\n", rootUse, rootUse)
				}
				for _, r := range results {
					fmt.Fprintf(out, "%s@%s\n", r.Source, r.SkillID)
				}
			default:
				return fmt.Errorf("unknown --format %q (expected: table|add)", format)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "Max results")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table or add. The add format outputs one skill per line: <source>@<skill_id>.")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress add-format hints on stderr")
	return cmd
}

func newAddCmd(out *os.File, rootOpts *RootOptions, enableAgentFlags bool, rootUse string) *cobra.Command {
	var global bool
	var agentNames []string
	var skillNames []string
	var listOnly bool
	var yes bool
	var all bool
	var fullDepth bool
	var mode string

	cmd := &cobra.Command{
		Use:     "add <source>",
		Args:    cobra.ExactArgs(1),
		Short:   "Install skills from local dir, git repo, URL, or well-known",
		Example: fmt.Sprintf("  %s add github.com/user/repo\n  %s add github.com/user/repo@skill-name\n  %s add -g --all github.com/user/repo", rootUse, rootUse, rootUse),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var agents []skills.AgentType
			if enableAgentFlags {
				agents = parseAgents(agentNames)
			}

			var dirs []string
			if global {
				if rootOpts.GlobalDir != "" {
					dirs = []string{rootOpts.GlobalDir}
				}
			} else {
				if rootOpts.WorkspaceDir != "" {
					dirs = []string{rootOpts.WorkspaceDir}
				}
			}

			enableAgentDiscovery := rootOpts.EnableAgentDiscovery
			if !enableAgentFlags {
				enableAgentDiscovery = false
				if len(dirs) == 0 {
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					canonical, err := skills.CanonicalSkillsDir(global, cwd)
					if err != nil {
						return err
					}
					dirs = []string{canonical}
				}
			}

			res, err := skills.Add(ctx, skills.AddOptions{
				Source:               args[0],
				Global:               global,
				Agents:               agents,
				Skills:               skillNames,
				ListOnly:             listOnly,
				Yes:                  yes,
				All:                  all,
				FullDepth:            fullDepth,
				Mode:                 skills.InstallMode(mode),
				Dirs:                 dirs,
				EnableAgentDiscovery: enableAgentDiscovery,
			})
			if err != nil {
				return err
			}

			if listOnly {
				for _, s := range res.Available {
					fmt.Fprintf(out, "%s\t%s\n", s.Name, s.Description)
				}
				return nil
			}

			for _, s := range res.Installed {
				fmt.Fprintf(out, "installed\t%s\t%s\n", s.InstallName, s.Name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&global, "global", "g", false, "Install globally")
	if enableAgentFlags {
		cmd.Flags().StringArrayVarP(&agentNames, "agent", "a", nil, "Target agents (repeatable, or '*')")
	}
	if enableAgentFlags {
		cmd.Flags().BoolVar(&all, "all", false, "Install all skills to all detected agents")
	}
	cmd.Flags().StringArrayVarP(&skillNames, "skill", "s", nil, "Skill names (repeatable, or '*')")
	cmd.Flags().BoolVarP(&listOnly, "list", "l", false, "List skills but do not install")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Non-interactive defaults")

	cmd.Flags().BoolVar(&fullDepth, "full-depth", false, "Deep-scan for SKILL.md files")
	cmd.Flags().StringVar(&mode, "mode", string(skills.InstallModeSymlink), "Install mode: symlink or copy")
	return cmd
}

// newListCmd creates the "list" command, which lists installed skills.
func newListCmd(out *os.File, rootOpts *RootOptions, enableAgentFlags bool, rootUse string) *cobra.Command {
	var global bool
	var agentNames []string

	cmd := &cobra.Command{
		Use:     "list",
		Args:    cobra.NoArgs,
		Short:   "List installed skills",
		Example: fmt.Sprintf("  %s list\n  %s list -g\n  %s list -a claude", rootUse, rootUse, rootUse),
		RunE: func(cmd *cobra.Command, args []string) error {
			var agents []skills.AgentType
			if enableAgentFlags {
				agents = parseAgents(agentNames)
			}
			items, err := listCmdItems(*rootOpts, agents, global, enableAgentFlags)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintf(out, "No skills installed, you can use add to install\n")
				return nil
			}
			for _, it := range items {
				fmt.Fprintln(out, it.Name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "List global installs only (default: workspace + global)")
	if enableAgentFlags {
		cmd.Flags().StringArrayVarP(&agentNames, "agent", "a", nil, "Filter by agent (repeatable, or '*')")
	}
	return cmd
}

// newGetCmd creates the "get" command, which prints the content of an installed skill.
func newGetCmd(out *os.File, rootOpts *RootOptions, enableAgentFlags bool, rootUse string) *cobra.Command {
	var global bool
	var agentNames []string

	cmd := &cobra.Command{
		Use:     "get <skill>",
		Args:    cobra.ExactArgs(1),
		Short:   "get the content of an installed skill",
		Example: fmt.Sprintf("  %s get my-skill\n  %s get -g my-skill", rootUse, rootUse),
		RunE: func(cmd *cobra.Command, args []string) error {
			var agents []skills.AgentType
			if enableAgentFlags {
				agents = parseAgents(agentNames)
			}
			res, err := getCmdResult(*rootOpts, args[0], agents, global, enableAgentFlags)
			if err != nil {
				return err
			}
			fmt.Fprint(out, res.Skill.Content)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Look in global installs only (default: workspace + global)")
	if enableAgentFlags {
		cmd.Flags().StringArrayVarP(&agentNames, "agent", "a", nil, "Filter by agent (repeatable, or '*')")
	}
	return cmd
}

// newRemoveCmd creates the "remove" command, which removes installed skills.
func newRemoveCmd(out *os.File, rootOpts *RootOptions, enableAgentFlags bool, rootUse string) *cobra.Command {
	var global bool
	var agentNames []string
	var all bool
	var yes bool

	cmd := &cobra.Command{
		Use:     "remove [skill...]",
		Args:    cobra.ArbitraryArgs,
		Short:   "Remove installed skills",
		Example: fmt.Sprintf("  %s remove my-skill\n  %s remove -g --all\n  %s remove -a claude my-skill", rootUse, rootUse, rootUse),
		RunE: func(cmd *cobra.Command, args []string) error {
			var agents []skills.AgentType
			if enableAgentFlags {
				agents = parseAgents(agentNames)
			}
			res, err := removeScoped(*rootOpts, args, agents, global, all, yes, enableAgentFlags)
			if err != nil {
				return err
			}
			for _, r := range res.Removed {
				fmt.Fprintf(out, "removed\t%s\n", r.InstallName)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&global, "global", "g", false, "Remove from global install")
	if enableAgentFlags {
		cmd.Flags().StringArrayVarP(&agentNames, "agent", "a", nil, "Target agents (repeatable, or '*')")
		cmd.Flags().BoolVar(&all, "all", false, "Remove all installed skills")
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Non-interactive defaults")
	return cmd
}

// newInitCmd creates the "init" command, which writes a SKILL.md template to disk.
func newInitCmd(out *os.File, rootUse string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init [name]",
		Args:    cobra.MaximumNArgs(1),
		Short:   "Create a SKILL.md template",
		Example: fmt.Sprintf("  %s init\n  %s init my-skill", rootUse, rootUse),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			path, err := skills.Init(skills.InitOptions{Name: name})
			if err != nil {
				return err
			}
			fmt.Fprintln(out, path)
			return nil
		},
	}
	return cmd
}

// newCheckCmd creates the "check" command, which checks for updates to globally installed skills.
func newCheckCmd(out *os.File, rootUse string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "check",
		Args:    cobra.NoArgs,
		Short:   "Check for updates of globally installed GitHub skills",
		Example: fmt.Sprintf("  %s check", rootUse),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			updates, err := skills.CheckUpdates(ctx)
			if err != nil {
				return err
			}
			for _, u := range updates {
				fmt.Fprintf(out, "%s\t%s\t%s\n", u.InstallName, u.CurrentHash, u.RemoteHash)
			}
			return nil
		},
	}
	return cmd
}

// newUpdateCmd creates the "update" command, which applies updates to globally installed skills.
func newUpdateCmd(rootUse string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Args:    cobra.NoArgs,
		Short:   "Update globally installed GitHub skills",
		Example: fmt.Sprintf("  %s update", rootUse),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			_, err := skills.Update(ctx)
			return err
		},
	}
	return cmd
}

func parseAgents(list []string) []skills.AgentType {
	if len(list) == 0 {
		return nil
	}
	for _, v := range list {
		if v == "*" {
			return skills.AllAgentTypes()
		}
	}
	out := make([]skills.AgentType, 0, len(list))
	for _, v := range list {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, skills.AgentType(v))
	}
	return out
}

func agentTypeStrings(list []skills.AgentType) []string {
	out := make([]string, 0, len(list))
	for _, a := range list {
		out = append(out, string(a))
	}
	return out
}

func listScoped(rootOpts RootOptions, agents []skills.AgentType, global bool) ([]skills.ListedSkill, error) {
	opts := skills.ListOptions{
		Global:               global,
		Agents:               agents,
		EnableAgentDiscovery: rootOpts.EnableAgentDiscovery,
	}
	if global {
		if rootOpts.GlobalDir != "" {
			opts.Dirs = []string{rootOpts.GlobalDir}
		}
	} else {
		if rootOpts.WorkspaceDir != "" {
			opts.Dirs = []string{rootOpts.WorkspaceDir}
		}
	}
	return skills.List(opts)
}

func getScoped(rootOpts RootOptions, skill string, agents []skills.AgentType, global bool) (skills.GetResult, error) {
	opts := skills.GetOptions{
		Skill:                skill,
		Global:               global,
		Agents:               agents,
		EnableAgentDiscovery: rootOpts.EnableAgentDiscovery,
	}
	if global {
		if rootOpts.GlobalDir != "" {
			opts.Dirs = []string{rootOpts.GlobalDir}
		}
	} else {
		if rootOpts.WorkspaceDir != "" {
			opts.Dirs = []string{rootOpts.WorkspaceDir}
		}
	}
	return skills.Get(opts)
}

func removeScoped(rootOpts RootOptions, skillsToRemove []string, agents []skills.AgentType, global bool, all bool, yes bool, enableAgentDirs bool) (skills.RemoveResult, error) {
	opts := skills.RemoveOptions{
		Global:               global,
		Agents:               agents,
		Skills:               skillsToRemove,
		All:                  all,
		Yes:                  yes,
		EnableAgentDiscovery: rootOpts.EnableAgentDiscovery,
	}
	if global {
		if rootOpts.GlobalDir != "" {
			opts.Dirs = []string{rootOpts.GlobalDir}
		}
	} else {
		if rootOpts.WorkspaceDir != "" {
			opts.Dirs = []string{rootOpts.WorkspaceDir}
		}
	}
	if len(opts.Dirs) == 0 && !enableAgentDirs {
		cwd, err := os.Getwd()
		if err != nil {
			return skills.RemoveResult{}, err
		}
		canonical, err := skills.CanonicalSkillsDir(global, cwd)
		if err != nil {
			return skills.RemoveResult{}, err
		}
		opts.Dirs = []string{canonical}
		opts.EnableAgentDiscovery = false
	}
	return skills.Remove(opts)
}

func listCmdItems(rootOpts RootOptions, agents []skills.AgentType, global bool, enableAgentDirs bool) ([]skills.ListedSkill, error) {
	if global {
		if enableAgentDirs {
			return listScoped(rootOpts, agents, true)
		}
		dir := rootOpts.GlobalDir
		if dir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			dir, err = skills.CanonicalSkillsDir(true, cwd)
			if err != nil {
				return nil, err
			}
		}
		return skills.List(skills.ListOptions{Dirs: []string{dir}})
	}

	dirs, agentList, err := resolveWorkspaceThenGlobalDirs(rootOpts, agents, enableAgentDirs)
	if err != nil {
		return nil, err
	}

	items, err := skills.List(skills.ListOptions{Dirs: dirs})
	if err != nil {
		return nil, err
	}

	if len(agentList) == 0 {
		return items, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	workspaceAgentDirs := map[skills.AgentType]string{}
	globalAgentDirs := map[skills.AgentType]string{}
	for _, a := range agentList {
		if dir, err := skills.AgentSkillsDir(a, false, cwd); err == nil {
			workspaceAgentDirs[a] = dir
		}
		if dir, err := skills.AgentSkillsDir(a, true, cwd); err == nil {
			globalAgentDirs[a] = dir
		}
	}

	for i := range items {
		var present []skills.AgentType
		for _, a := range agentList {
			if dir, ok := workspaceAgentDirs[a]; ok {
				if _, err := os.Lstat(filepath.Join(dir, items[i].InstallName)); err == nil {
					present = append(present, a)
					continue
				}
			}
			if dir, ok := globalAgentDirs[a]; ok {
				if _, err := os.Lstat(filepath.Join(dir, items[i].InstallName)); err == nil {
					present = append(present, a)
					continue
				}
			}
		}
		items[i].Agents = present
	}

	return items, nil
}

func getCmdResult(rootOpts RootOptions, skill string, agents []skills.AgentType, global bool, enableAgentDirs bool) (skills.GetResult, error) {
	if global {
		if enableAgentDirs {
			return getScoped(rootOpts, skill, agents, true)
		}

		dir := rootOpts.GlobalDir
		if dir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return skills.GetResult{}, err
			}
			dir, err = skills.CanonicalSkillsDir(true, cwd)
			if err != nil {
				return skills.GetResult{}, err
			}
		}
		return skills.Get(skills.GetOptions{Skill: skill, Global: true, Dirs: []string{dir}})
	}

	dirs, _, err := resolveWorkspaceThenGlobalDirs(rootOpts, agents, enableAgentDirs)
	if err != nil {
		return skills.GetResult{}, err
	}
	return skills.Get(skills.GetOptions{Skill: skill, Dirs: dirs})
}

func resolveWorkspaceThenGlobalDirs(rootOpts RootOptions, agents []skills.AgentType, includeAgentDirs bool) ([]string, []skills.AgentType, error) {
	var dirs []string
	if rootOpts.GlobalDir != "" {
		dirs = append(dirs, rootOpts.GlobalDir)
	}
	if rootOpts.WorkspaceDir != "" {
		dirs = append(dirs, rootOpts.WorkspaceDir)
	}
	if len(dirs) > 0 {
		return dirs, nil, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	globalCanonical, err := skills.CanonicalSkillsDir(true, cwd)
	if err != nil {
		return nil, nil, err
	}
	workspaceCanonical, err := skills.CanonicalSkillsDir(false, cwd)
	if err != nil {
		return nil, nil, err
	}

	dirs = append(dirs, globalCanonical)
	dirs = append(dirs, workspaceCanonical)
	if !includeAgentDirs {
		return dirs, nil, nil
	}

	if !rootOpts.EnableAgentDiscovery {
		return nil, nil, skills.ErrAgentDiscoveryRequired
	}

	agentList := agents
	if len(agentList) == 0 {
		agentList, err = skills.DetectInstalledAgents(cwd)
		if err != nil {
			return nil, nil, err
		}
	}

	for _, a := range agentList {
		dir, err := skills.AgentSkillsDir(a, true, cwd)
		if err != nil {
			continue
		}
		dirs = append(dirs, dir)
	}
	for _, a := range agentList {
		dir, err := skills.AgentSkillsDir(a, false, cwd)
		if err != nil {
			continue
		}
		dirs = append(dirs, dir)
	}

	return dirs, agentList, nil
}

// WithContext sets ctx on cmd and returns cmd for fluent chaining.
func WithContext(ctx context.Context, cmd *cobra.Command) *cobra.Command {
	cmd.SetContext(ctx)
	return cmd
}
