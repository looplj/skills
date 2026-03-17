package skillscmd

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	skills "github.com/looplj/skills"
)

func TestNewRootCommand_CommandsFilter(t *testing.T) {
	root := NewRootCommand(RootOptions{
		Use:                  "skills",
		Commands:             []string{"search"},
		EnableAgentDiscovery: false,
	})

	if cmd := findSubcommand(root, "search"); cmd == nil {
		t.Fatalf("expected search command")
	}
	if cmd := findSubcommand(root, "add"); cmd != nil {
		t.Fatalf("did not expect add command")
	}
	if cmd := findSubcommand(root, "list"); cmd != nil {
		t.Fatalf("did not expect list command")
	}
	if cmd := findSubcommand(root, "get"); cmd != nil {
		t.Fatalf("did not expect get command")
	}
	if cmd := findSubcommand(root, "remove"); cmd != nil {
		t.Fatalf("did not expect remove command")
	}
	if cmd := findSubcommand(root, "init"); cmd != nil {
		t.Fatalf("did not expect init command")
	}
	if cmd := findSubcommand(root, "check"); cmd != nil {
		t.Fatalf("did not expect check command")
	}
	if cmd := findSubcommand(root, "update"); cmd != nil {
		t.Fatalf("did not expect update command")
	}
}

func TestNewRootCommand_EnableAgentFlags_DefaultTrue(t *testing.T) {
	root := NewRootCommand(RootOptions{
		Use:                  "skills",
		EnableAgentDiscovery: true,
		EnableAgentFlags:     true,
	})

	assertFlagExists(t, findSubcommand(root, "add"), "agent")
	assertFlagExists(t, findSubcommand(root, "add"), "all")
	assertFlagExists(t, findSubcommand(root, "list"), "agent")
	assertFlagExists(t, findSubcommand(root, "get"), "agent")
	assertFlagExists(t, findSubcommand(root, "remove"), "agent")
	assertFlagExists(t, findSubcommand(root, "remove"), "all")
}

func TestNewRootCommand_EnableAgentFlags_Disabled(t *testing.T) {
	enableAgentFlags := false
	root := NewRootCommand(RootOptions{
		Use:                  "skills",
		EnableAgentDiscovery: true,
		EnableAgentFlags:     enableAgentFlags,
	})

	assertFlagMissing(t, findSubcommand(root, "add"), "agent")
	assertFlagMissing(t, findSubcommand(root, "add"), "all")
	assertFlagMissing(t, findSubcommand(root, "list"), "agent")
	assertFlagMissing(t, findSubcommand(root, "get"), "agent")
	assertFlagMissing(t, findSubcommand(root, "remove"), "agent")
	assertFlagMissing(t, findSubcommand(root, "remove"), "all")
}

func TestNewRootCommand_BundledSkillsFuncRunsInPersistentPreRun(t *testing.T) {
	called := false
	root := NewRootCommand(RootOptions{
		Use: "skills",
		BundledSkillsFunc: func(ctx context.Context) []skills.Skill {
			if ctx == nil {
				t.Fatalf("expected context")
			}
			called = true
			return []skills.Skill{{Name: "seo-audit", Description: "bundled"}}
		},
	})

	root.SetContext(context.Background())
	root.PersistentPreRun(root, nil)
	if !called {
		t.Fatalf("expected BundledSkillsFunc to be called")
	}
}

func findSubcommand(root *cobra.Command, name string) *cobra.Command {
	for _, c := range root.Commands() {
		if c == nil {
			continue
		}
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func assertFlagExists(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()
	if cmd == nil {
		t.Fatalf("expected command for flag %q", name)
	}
	if cmd.Flags().Lookup(name) == nil {
		t.Fatalf("expected flag %q on command %q", name, cmd.Name())
	}
}

func assertFlagMissing(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()
	if cmd == nil {
		t.Fatalf("expected command for flag %q", name)
	}
	if cmd.Flags().Lookup(name) != nil {
		t.Fatalf("did not expect flag %q on command %q", name, cmd.Name())
	}
}
