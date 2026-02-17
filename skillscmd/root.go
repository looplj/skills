package skillscmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	skills "github.com/looplj/skills"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	Use                  string
	Stdout               *os.File
	Stderr               *os.File
	EnableAgentDiscovery bool
	Dirs                 []string
}

func NewRootCommand(opts RootOptions) *cobra.Command {
	if opts.Use == "" {
		opts.Use = "skills"
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	root := &cobra.Command{
		Use:           opts.Use,
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Manage AI agent skills",
	}

	root.AddCommand(newFindCmd(stdout))
	root.AddCommand(newAddCmd(stdout, opts))
	root.AddCommand(newListCmd(stdout, opts))
	root.AddCommand(newShowCmd(stdout, opts))
	root.AddCommand(newRemoveCmd(stdout, opts))
	root.AddCommand(newInitCmd(stdout))
	root.AddCommand(newCheckCmd(stdout))
	root.AddCommand(newUpdateCmd(stdout))

	root.SetOut(stdout)
	root.SetErr(stderr)

	return root
}

func newFindCmd(out *os.File) *cobra.Command {
	var limit int
	var format string
	cmd := &cobra.Command{
		Use:     "find <query>",
		Args:    cobra.MinimumNArgs(1),
		Short:   "Search skills registry",
		Example: "  skills find docker\n  skills find --limit 5 kubernetes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			results, err := skills.Find(ctx, strings.Join(args, " "), limit)
			if err != nil {
				return err
			}
			switch format {
			case "", "table":
				for _, r := range results {
					fmt.Fprintf(out, "%s\t%s\t%s\t%d\n", r.Name, r.ID, r.Source, r.Installs)
				}
			case "add":
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
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table or add")
	return cmd
}

func newAddCmd(out *os.File, rootOpts RootOptions) *cobra.Command {
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
		Example: "  skills add ./my-skill\n  skills add github.com/user/repo\n  skills add github.com/user/repo@skill-name\n  skills add -g --all github.com/user/repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			agents := parseAgents(agentNames)

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
				Dirs:                 rootOpts.Dirs,
				EnableAgentDiscovery: rootOpts.EnableAgentDiscovery,
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

	cmd.Flags().BoolVarP(&global, "global", "g", false, "Install globally (~/.agents/skills)")
	cmd.Flags().StringArrayVarP(&agentNames, "agent", "a", nil, "Target agents (repeatable, or '*')")
	cmd.Flags().StringArrayVarP(&skillNames, "skill", "s", nil, "Skill names (repeatable, or '*')")
	cmd.Flags().BoolVarP(&listOnly, "list", "l", false, "List skills but do not install")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Non-interactive defaults")
	cmd.Flags().BoolVar(&all, "all", false, "Install all skills to all detected agents")
	cmd.Flags().BoolVar(&fullDepth, "full-depth", false, "Deep-scan for SKILL.md files")
	cmd.Flags().StringVar(&mode, "mode", string(skills.InstallModeSymlink), "Install mode: symlink or copy")
	return cmd
}

func newListCmd(out *os.File, rootOpts RootOptions) *cobra.Command {
	var global bool
	var agentNames []string

	cmd := &cobra.Command{
		Use:     "list",
		Args:    cobra.NoArgs,
		Short:   "List installed skills",
		Example: "  skills list\n  skills list -g\n  skills list -a claude",
		RunE: func(cmd *cobra.Command, args []string) error {
			agents := parseAgents(agentNames)
			items, err := skills.List(skills.ListOptions{
				Global:               global,
				Agents:               agents,
				Dirs:                 rootOpts.Dirs,
				EnableAgentDiscovery: rootOpts.EnableAgentDiscovery,
			})
			if err != nil {
				return err
			}
			for _, it := range items {
				fmt.Fprintf(out, "%s\t%s\t%s\n", it.InstallName, it.Name, strings.Join(agentTypeStrings(it.Agents), ","))
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "List global installs")
	cmd.Flags().StringArrayVarP(&agentNames, "agent", "a", nil, "Filter by agent (repeatable, or '*')")
	return cmd
}

func newShowCmd(out *os.File, rootOpts RootOptions) *cobra.Command {
	var global bool
	var agentNames []string

	cmd := &cobra.Command{
		Use:     "get <skill>",
		Args:    cobra.ExactArgs(1),
		Short:   "get the content of an installed skill",
		Example: "  skills get my-skill\n  skills get -g my-skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			agents := parseAgents(agentNames)
			res, err := skills.Get(skills.GetOptions{
				Skill:                args[0],
				Global:               global,
				Agents:               agents,
				Dirs:                 rootOpts.Dirs,
				EnableAgentDiscovery: rootOpts.EnableAgentDiscovery,
			})
			if err != nil {
				return err
			}
			fmt.Fprint(out, res.Skill.Content)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Look in global installs")
	cmd.Flags().StringArrayVarP(&agentNames, "agent", "a", nil, "Filter by agent (repeatable, or '*')")
	return cmd
}

func newRemoveCmd(out *os.File, rootOpts RootOptions) *cobra.Command {
	var global bool
	var agentNames []string
	var all bool
	var yes bool

	cmd := &cobra.Command{
		Use:     "remove [skill...]",
		Args:    cobra.ArbitraryArgs,
		Short:   "Remove installed skills",
		Example: "  skills remove my-skill\n  skills remove -g --all\n  skills remove -a claude my-skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			agents := parseAgents(agentNames)
			res, err := skills.Remove(skills.RemoveOptions{
				Global:               global,
				Agents:               agents,
				Skills:               args,
				All:                  all,
				Yes:                  yes,
				Dirs:                 rootOpts.Dirs,
				EnableAgentDiscovery: rootOpts.EnableAgentDiscovery,
			})
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
	cmd.Flags().StringArrayVarP(&agentNames, "agent", "a", nil, "Target agents (repeatable, or '*')")
	cmd.Flags().BoolVar(&all, "all", false, "Remove all installed skills")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Non-interactive defaults")
	return cmd
}

func newInitCmd(out *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init [name]",
		Args:    cobra.MaximumNArgs(1),
		Short:   "Create a SKILL.md template",
		Example: "  skills init\n  skills init my-skill",
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

func newCheckCmd(out *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "check",
		Args:    cobra.NoArgs,
		Short:   "Check for updates of globally installed GitHub skills",
		Example: "  skills check",
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

func newUpdateCmd(out *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Args:    cobra.NoArgs,
		Short:   "Update globally installed GitHub skills",
		Example: "  skills update",
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

func WithContext(ctx context.Context, cmd *cobra.Command) *cobra.Command {
	cmd.SetContext(ctx)
	return cmd
}
