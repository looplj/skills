package main

import (
	"fmt"
	"os"

	skillscmd "github.com/looplj/skills/skillscmd"
)

func main() {
	cmd := skillscmd.NewRootCommand(skillscmd.RootOptions{
		Use:                  "find-skills",
		Stdout:               os.Stdout,
		Stderr:               os.Stderr,
		EnableAgentDiscovery: true,
	})
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
