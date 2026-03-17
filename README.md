# skills (Go 移植版)

本项目是对 Vercel 的技能管理 CLI 的 Go 语言移植实现，主要用于发现、安装、列出与移除 Agent Skills（基于 `SKILL.md`）。

原项目：
- https://github.com/vercel-labs/skills （上游实现，Node.js/TypeScript）

## 安装

安装到全局（推荐）：

```bash
go install github.com/looplj/skills/cmd/find-skills@latest
```

安装后可直接使用：

```bash
find-skills --help
```

从源码构建：

```bash
go build -o find-skills ./cmd/find-skills
```

或直接运行：

```bash
go run ./cmd/find-skills --help
```

## 快速开始

搜索技能：

```bash
find-skills find "browser automation" --limit 10
```

以可直接用于安装的格式输出搜索结果：

```bash
find-skills find "pdf" --format add
```

安装技能（支持本地目录、Git 仓库、URL、well-known 等来源）：

```bash
find-skills add vercel-labs/agent-skills --global --yes
```

列出已安装技能：

```bash
find-skills list --global
```

移除技能：

```bash
find-skills remove --global my-skill
```

生成 `SKILL.md` 模板：

```bash
find-skills init MySkill
```

检查并更新通过 GitHub 安装的全局技能：

```bash
find-skills check
find-skills update
```

## 作为库使用

该 Go 版本不仅提供可执行文件，也可以直接作为库引入，在你的程序里完成技能的发现/安装/移除等逻辑。

安装技能（启用 Agent 自动发现；若需安装到指定目录，见下一个示例）：

```go
package main

import (
	"context"
	"fmt"

	skills "github.com/looplj/skills"
)

func main() {
	ctx := context.Background()
	res, err := skills.Add(ctx, skills.AddOptions{
		Source:               "vercel-labs/agent-skills",
		Global:               true,
		Yes:                  true,
		EnableAgentDiscovery: true,
	})
	if err != nil {
		panic(err)
	}
	for _, s := range res.Installed {
		fmt.Println(s.InstallName, s.Name)
	}
}
```

安装技能（显式指定目标目录；适合嵌入到你自己的目录结构中）：

```go
package main

import (
	"context"

	skills "github.com/looplj/skills"
)

func main() {
	ctx := context.Background()
	_, err := skills.Add(ctx, skills.AddOptions{
		Source:               "vercel-labs/agent-skills",
		Yes:                  true,
		Dirs:                 []string{".agents/skills"},
		EnableAgentDiscovery: false,
	})
	if err != nil {
		panic(err)
	}
}
```

列出与移除：

```go
package main

import (
	"fmt"

	skills "github.com/looplj/skills"
)

func main() {
	items, err := skills.List(skills.ListOptions{
		Global:               true,
		EnableAgentDiscovery: true,
	})
	if err != nil {
		panic(err)
	}
	for _, it := range items {
		fmt.Println(it.InstallName, it.Name)
	}
}
```

```go
package main

import (
	skills "github.com/looplj/skills"
)

func main() {
	_, err := skills.Remove(skills.RemoveOptions{
		Global:               true,
		Skills:               []string{"my-skill"},
		EnableAgentDiscovery: true,
	})
	if err != nil {
		panic(err)
	}
}
```

## 可嵌入 CLI（Cobra）

项目同时提供了可嵌入的 Cobra 命令集合，可作为你自己 CLI 的子命令使用。

将 `find-skills` 作为子命令挂到你的根命令下：

```go
package main

import (
	"os"

	skillscmd "github.com/looplj/skills/skillscmd"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "myapp"}
	root.AddCommand(skillscmd.NewRootCommand(skillscmd.RootOptions{
		Use:                  "find-skills",
		Stdout:               os.Stdout,
		Stderr:               os.Stderr,
		EnableAgentDiscovery: true,
	}))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
```

如果你不希望依赖自动发现，也可以通过 `WorkspaceDir` / `GlobalDir` 指定技能目录（安装/列出/移除会直接作用在这些目录上）：

```go
package main

import (
	"os"

	skillscmd "github.com/looplj/skills/skillscmd"
)

func main() {
	cmd := skillscmd.NewRootCommand(skillscmd.RootOptions{
		Use:                  "find-skills",
		Stdout:               os.Stdout,
		Stderr:               os.Stderr,
		WorkspaceDir:         ".agents/skills",
		GlobalDir:            "~/.agents/skills",
		EnableAgentDiscovery: false,
	})
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

如果你还需要暴露一层 bundled skills 作为 `list/get` 的回退来源，可以直接从代码注入 `BundledSkills`。`Skill.Name` 会作为 install name 使用。当 bundled skill 与已安装 skill 同名时，已安装的 global/workspace skill 会覆盖 bundled 版本：

```go
root.AddCommand(skillscmd.NewRootCommand(skillscmd.RootOptions{
	Use:          "find-skills",
	WorkspaceDir: ".agents/skills",
	GlobalDir:    "~/.agents/skills",
	BundledSkills: []skills.Skill{
		{
			Name:        "seo-audit",
			Description: "Bundled SEO audit skill",
			Content:     "---\nname: seo-audit\ndescription: Bundled SEO audit skill\n---\n",
		},
	},
}))
```

## 认证

在检查/更新 GitHub 技能时，若遇到限流或私有仓库访问需求，可配置以下任一方式：
- 环境变量 `GITHUB_TOKEN` 或 `GH_TOKEN`
- 已登录的 GitHub CLI（`gh auth login`），工具会尝试读取 `gh auth token`

## 许可证

MIT，见 [LICENSE](./LICENSE)。本项目保持与上游项目声明一致的开源协议。
