package skills

import (
	"errors"
)

var (
	ErrSourceRequired         = errors.New("source is required")
	ErrAgentDiscoveryRequired = errors.New("target dirs are required when agent discovery is disabled")
	ErrNoTargetDirs           = errors.New("no target dirs specified")
	ErrNoSkillsFound          = errors.New("no skills found")
	ErrNoSkillsSpecified      = errors.New("no skills specified")
	ErrSkillNameRequired      = errors.New("skill name is required")
	ErrMultipleSkills         = errors.New("multiple skills found; specify --skill or --yes")
	ErrMultipleWellKnown      = errors.New("multiple well-known skills found; specify --skill or --yes")
	ErrUnsupportedSourceType  = errors.New("unsupported source type")
	ErrUnknownInstallMode     = errors.New("unknown install mode")
	ErrUnrecognizedSource     = errors.New("unrecognized source format")
	ErrMissingFrontmatter     = errors.New("missing frontmatter")
	ErrMissingName            = errors.New("frontmatter missing name")
	ErrMissingDescription     = errors.New("frontmatter missing description")
	ErrHomeNotFound           = errors.New("home directory not found")
	ErrNoWellKnownIndex       = errors.New("no well-known skills index found")
	ErrUnsafeFilePath         = errors.New("unsafe file path in well-known index")
	ErrSkillAlreadyExists     = errors.New("SKILL.md already exists")
	ErrSkillNotFound          = errors.New("skill not found")
)

// func ErrSkillNotFound(name string) error {
// 	return fmt.Errorf("skill not found: %s", name)
// }
