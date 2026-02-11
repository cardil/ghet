package tasks

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cardil/ghet/pkg/ghet/download"
	"github.com/cardil/ghet/pkg/ghet/install"
	"github.com/cardil/ghet/pkg/github"
	"github.com/goyek/goyek/v2"
)

func Lint() goyek.Task {
	return goyek.Task{
		Name:  "lint",
		Usage: "Run linters",
		Action: func(a *goyek.A) {
			// Run editorconfig-checker
			runTool(a, toolConfig{
				envVar:  "EDITORCONFIG_CHECKER_VERSION",
				version: "v3.1.0",
				owner:   "editorconfig-checker",
				repo:    "editorconfig-checker",
				args:    nil,
			})

			// Run golangci-lint
			runTool(a, toolConfig{
				envVar:  "GOLANGCI_LINT_VERSION",
				version: "v2.4.0",
				owner:   "golangci",
				repo:    "golangci-lint",
				args:    []string{"run", "./..."},
			})
		},
	}
}

type toolConfig struct {
	envVar  string
	version string
	owner   string
	repo    string
	args    []string
}

func runTool(a *goyek.A, cfg toolConfig) {
	version := os.Getenv(cfg.envVar)
	if version == "" {
		version = cfg.version
	}

	toolsDir := filepath.Join("build", "_output", "tools", cfg.repo+"-"+version)
	toolPath := filepath.Join(toolsDir, cfg.repo)

	if _, err := os.Stat(toolPath); err != nil {
		if !os.IsNotExist(err) {
			a.Fatalf("checking %s/%s path: %v", cfg.owner, cfg.repo, err)
		}
		a.Logf("Downloading %s/%s %s", cfg.owner, cfg.repo, version)
		args := download.Args{
			Args: install.Args{
				Asset: github.Asset{
					Release: github.Release{
						Tag: version,
						Repository: github.Repository{
							Owner: cfg.owner,
							Repo:  cfg.repo,
						},
					},
				},
			},
			Destination: toolsDir,
		}
		args.Args = args.Args.WithDefaults()

		if err := download.Action(a.Context(), args); err != nil {
			a.Fatal(err)
		}
	}

	cmd := exec.CommandContext(a.Context(), toolPath, cfg.args...)
	cmd.Stdout = a.Output()
	cmd.Stderr = a.Output()
	if err := cmd.Run(); err != nil {
		a.Fatal(err)
	}
}
