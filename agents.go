package skills

import (
	"os"
	"path/filepath"
)

var allAgentTypes = []AgentType{
	"amp",
	"antigravity",
	"augment",
	"claude-code",
	"openclaw",
	"cline",
	"codebuddy",
	"codex",
	"command-code",
	"continue",
	"crush",
	"cursor",
	"droid",
	"gemini-cli",
	"github-copilot",
	"goose",
	"junie",
	"iflow-cli",
	"kilo",
	"kimi-cli",
	"kiro-cli",
	"kode",
	"mcpjam",
	"mistral-vibe",
	"mux",
	"opencode",
	"openhands",
	"pi",
	"qoder",
	"qwen-code",
	"replit",
	"roo",
	"trae",
	"trae-cn",
	"windsurf",
	"zencoder",
	"neovate",
	"pochi",
	"adal",
}

func AllAgentTypes() []AgentType {
	out := make([]AgentType, 0, len(allAgentTypes))
	out = append(out, allAgentTypes...)

	return out
}

func AgentConfigFor(agent AgentType) (AgentConfig, error) {
	home, err := HomeDir()
	if err != nil {
		return AgentConfig{}, err
	}

	configHome, err := XDGConfigHome()
	if err != nil {
		return AgentConfig{}, err
	}

	codexHome, err := CodexHome()
	if err != nil {
		return AgentConfig{}, err
	}

	claudeHome, err := ClaudeHome()
	if err != nil {
		return AgentConfig{}, err
	}

	switch agent {
	case "amp":
		return AgentConfig{Name: agent, DisplayName: "Amp", ProjectSkillsDir: ".agents/skills", GlobalSkillsDir: filepath.Join(configHome, "agents/skills"), ShowInUniversal: true, DetectGlobalPath: filepath.Join(configHome, "amp")}, nil
	case "antigravity":
		return AgentConfig{Name: agent, DisplayName: "Antigravity", ProjectSkillsDir: ".agent/skills", GlobalSkillsDir: filepath.Join(home, ".gemini/antigravity/skills"), ShowInUniversal: false, DetectProjectPath: ".agent", DetectGlobalPath: filepath.Join(home, ".gemini/antigravity")}, nil
	case "augment":
		return AgentConfig{Name: agent, DisplayName: "Augment", ProjectSkillsDir: ".augment/skills", GlobalSkillsDir: filepath.Join(home, ".augment/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".augment")}, nil
	case "claude-code":
		return AgentConfig{Name: agent, DisplayName: "Claude Code", ProjectSkillsDir: ".claude/skills", GlobalSkillsDir: filepath.Join(claudeHome, "skills"), ShowInUniversal: false, DetectGlobalPath: claudeHome}, nil
	case "openclaw":
		global := filepath.Join(home, ".moltbot/skills")
		if _, err := os.Stat(filepath.Join(home, ".openclaw")); err == nil {
			global = filepath.Join(home, ".openclaw/skills")
		} else if _, err := os.Stat(filepath.Join(home, ".clawdbot")); err == nil {
			global = filepath.Join(home, ".clawdbot/skills")
		}

		return AgentConfig{Name: agent, DisplayName: "OpenClaw", ProjectSkillsDir: "skills", GlobalSkillsDir: global, ShowInUniversal: false, DetectGlobalPath: ""}, nil
	case "cline":
		return AgentConfig{Name: agent, DisplayName: "Cline", ProjectSkillsDir: ".cline/skills", GlobalSkillsDir: filepath.Join(home, ".cline/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".cline")}, nil
	case "codebuddy":
		return AgentConfig{Name: agent, DisplayName: "CodeBuddy", ProjectSkillsDir: ".codebuddy/skills", GlobalSkillsDir: filepath.Join(home, ".codebuddy/skills"), ShowInUniversal: false, DetectProjectPath: ".codebuddy", DetectGlobalPath: filepath.Join(home, ".codebuddy")}, nil
	case "codex":
		return AgentConfig{Name: agent, DisplayName: "Codex", ProjectSkillsDir: ".agents/skills", GlobalSkillsDir: filepath.Join(codexHome, "skills"), ShowInUniversal: true, DetectGlobalPath: codexHome}, nil
	case "command-code":
		return AgentConfig{Name: agent, DisplayName: "Command Code", ProjectSkillsDir: ".commandcode/skills", GlobalSkillsDir: filepath.Join(home, ".commandcode/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".commandcode")}, nil
	case "continue":
		return AgentConfig{Name: agent, DisplayName: "Continue", ProjectSkillsDir: ".continue/skills", GlobalSkillsDir: filepath.Join(home, ".continue/skills"), ShowInUniversal: false, DetectProjectPath: ".continue", DetectGlobalPath: filepath.Join(home, ".continue")}, nil
	case "crush":
		return AgentConfig{Name: agent, DisplayName: "Crush", ProjectSkillsDir: ".crush/skills", GlobalSkillsDir: filepath.Join(home, ".config/crush/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".config/crush")}, nil
	case "cursor":
		return AgentConfig{Name: agent, DisplayName: "Cursor", ProjectSkillsDir: ".cursor/skills", GlobalSkillsDir: filepath.Join(home, ".cursor/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".cursor")}, nil
	case "droid":
		return AgentConfig{Name: agent, DisplayName: "Droid", ProjectSkillsDir: ".factory/skills", GlobalSkillsDir: filepath.Join(home, ".factory/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".factory")}, nil
	case "gemini-cli":
		return AgentConfig{Name: agent, DisplayName: "Gemini CLI", ProjectSkillsDir: ".agents/skills", GlobalSkillsDir: filepath.Join(home, ".gemini/skills"), ShowInUniversal: true, DetectGlobalPath: filepath.Join(home, ".gemini")}, nil
	case "github-copilot":
		return AgentConfig{Name: agent, DisplayName: "GitHub Copilot", ProjectSkillsDir: ".agents/skills", GlobalSkillsDir: filepath.Join(home, ".copilot/skills"), ShowInUniversal: true, DetectProjectPath: ".github", DetectGlobalPath: filepath.Join(home, ".copilot")}, nil
	case "goose":
		return AgentConfig{Name: agent, DisplayName: "Goose", ProjectSkillsDir: ".goose/skills", GlobalSkillsDir: filepath.Join(configHome, "goose/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(configHome, "goose")}, nil
	case "junie":
		return AgentConfig{Name: agent, DisplayName: "Junie", ProjectSkillsDir: ".junie/skills", GlobalSkillsDir: filepath.Join(home, ".junie/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".junie")}, nil
	case "iflow-cli":
		return AgentConfig{Name: agent, DisplayName: "iFlow CLI", ProjectSkillsDir: ".iflow/skills", GlobalSkillsDir: filepath.Join(home, ".iflow/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".iflow")}, nil
	case "kilo":
		return AgentConfig{Name: agent, DisplayName: "Kilo Code", ProjectSkillsDir: ".kilocode/skills", GlobalSkillsDir: filepath.Join(home, ".kilocode/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".kilocode")}, nil
	case "kimi-cli":
		return AgentConfig{Name: agent, DisplayName: "Kimi Code CLI", ProjectSkillsDir: ".agents/skills", GlobalSkillsDir: filepath.Join(home, ".config/agents/skills"), ShowInUniversal: true, DetectGlobalPath: filepath.Join(home, ".kimi")}, nil
	case "kiro-cli":
		return AgentConfig{Name: agent, DisplayName: "Kiro CLI", ProjectSkillsDir: ".kiro/skills", GlobalSkillsDir: filepath.Join(home, ".kiro/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".kiro")}, nil
	case "kode":
		return AgentConfig{Name: agent, DisplayName: "Kode", ProjectSkillsDir: ".kode/skills", GlobalSkillsDir: filepath.Join(home, ".kode/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".kode")}, nil
	case "mcpjam":
		return AgentConfig{Name: agent, DisplayName: "MCPJam", ProjectSkillsDir: ".mcpjam/skills", GlobalSkillsDir: filepath.Join(home, ".mcpjam/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".mcpjam")}, nil
	case "mistral-vibe":
		return AgentConfig{Name: agent, DisplayName: "Mistral Vibe", ProjectSkillsDir: ".vibe/skills", GlobalSkillsDir: filepath.Join(home, ".vibe/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".vibe")}, nil
	case "mux":
		return AgentConfig{Name: agent, DisplayName: "Mux", ProjectSkillsDir: ".mux/skills", GlobalSkillsDir: filepath.Join(home, ".mux/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".mux")}, nil
	case "opencode":
		return AgentConfig{Name: agent, DisplayName: "OpenCode", ProjectSkillsDir: ".agents/skills", GlobalSkillsDir: filepath.Join(configHome, "opencode/skills"), ShowInUniversal: true, DetectGlobalPath: filepath.Join(configHome, "opencode")}, nil
	case "openhands":
		return AgentConfig{Name: agent, DisplayName: "OpenHands", ProjectSkillsDir: ".openhands/skills", GlobalSkillsDir: filepath.Join(home, ".openhands/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".openhands")}, nil
	case "pi":
		return AgentConfig{Name: agent, DisplayName: "Pi", ProjectSkillsDir: ".pi/skills", GlobalSkillsDir: filepath.Join(home, ".pi/agent/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".pi/agent")}, nil
	case "qoder":
		return AgentConfig{Name: agent, DisplayName: "Qoder", ProjectSkillsDir: ".qoder/skills", GlobalSkillsDir: filepath.Join(home, ".qoder/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".qoder")}, nil
	case "qwen-code":
		return AgentConfig{Name: agent, DisplayName: "Qwen Code", ProjectSkillsDir: ".qwen/skills", GlobalSkillsDir: filepath.Join(home, ".qwen/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".qwen")}, nil
	case "replit":
		return AgentConfig{Name: agent, DisplayName: "Replit", ProjectSkillsDir: ".agents/skills", GlobalSkillsDir: filepath.Join(configHome, "agents/skills"), ShowInUniversal: false, DetectProjectPath: ".agents", DetectGlobalPath: ""}, nil
	case "roo":
		return AgentConfig{Name: agent, DisplayName: "Roo Code", ProjectSkillsDir: ".roo/skills", GlobalSkillsDir: filepath.Join(home, ".roo/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".roo")}, nil
	case "trae":
		return AgentConfig{Name: agent, DisplayName: "Trae", ProjectSkillsDir: ".trae/skills", GlobalSkillsDir: filepath.Join(home, ".trae/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".trae")}, nil
	case "trae-cn":
		return AgentConfig{Name: agent, DisplayName: "Trae CN", ProjectSkillsDir: ".trae/skills", GlobalSkillsDir: filepath.Join(home, ".trae-cn/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".trae-cn")}, nil
	case "windsurf":
		return AgentConfig{Name: agent, DisplayName: "Windsurf", ProjectSkillsDir: ".windsurf/skills", GlobalSkillsDir: filepath.Join(home, ".codeium/windsurf/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".codeium/windsurf")}, nil
	case "zencoder":
		return AgentConfig{Name: agent, DisplayName: "Zencoder", ProjectSkillsDir: ".zencoder/skills", GlobalSkillsDir: filepath.Join(home, ".zencoder/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".zencoder")}, nil
	case "neovate":
		return AgentConfig{Name: agent, DisplayName: "Neovate", ProjectSkillsDir: ".neovate/skills", GlobalSkillsDir: filepath.Join(home, ".neovate/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".neovate")}, nil
	case "pochi":
		return AgentConfig{Name: agent, DisplayName: "Pochi", ProjectSkillsDir: ".pochi/skills", GlobalSkillsDir: filepath.Join(home, ".pochi/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".pochi")}, nil
	case "adal":
		return AgentConfig{Name: agent, DisplayName: "AdaL", ProjectSkillsDir: ".adal/skills", GlobalSkillsDir: filepath.Join(home, ".adal/skills"), ShowInUniversal: false, DetectGlobalPath: filepath.Join(home, ".adal")}, nil
	default:
		return AgentConfig{}, os.ErrNotExist
	}
}

func DetectInstalledAgents(cwd string) ([]AgentType, error) {
	if cwd == "" {
		var err error

		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	var out []AgentType

	for _, a := range allAgentTypes {
		cfg, err := AgentConfigFor(a)
		if err != nil {
			continue
		}

		installed := false

		if cfg.DetectProjectPath != "" {
			if _, err := os.Stat(filepath.Join(cwd, cfg.DetectProjectPath)); err == nil {
				installed = true
			}
		}

		if !installed && cfg.DetectGlobalPath != "" {
			if _, err := os.Stat(cfg.DetectGlobalPath); err == nil {
				installed = true
			}
		}

		if !installed && a == "codex" {
			if _, err := os.Stat("/etc/codex"); err == nil {
				installed = true
			}
		}

		if installed {
			out = append(out, a)
		}
	}

	return out, nil
}

func AgentSkillsDir(agent AgentType, global bool, cwd string) (string, error) {
	cfg, err := AgentConfigFor(agent)
	if err != nil {
		return "", err
	}

	if global {
		return cfg.GlobalSkillsDir, nil
	}

	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(cwd, cfg.ProjectSkillsDir), nil
}

func IsUniversalAgent(agent AgentType) (bool, error) {
	cfg, err := AgentConfigFor(agent)
	if err != nil {
		return false, err
	}

	return cfg.ProjectSkillsDir == ".agents/skills", nil
}
