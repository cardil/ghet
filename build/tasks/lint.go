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
			runEditorconfigChecker(a)

			// Run golangci-lint
			runGolangciLint(a)
		},
	}
}

func runEditorconfigChecker(a *goyek.A) {
	version := os.Getenv("EDITORCONFIG_CHECKER_VERSION")
	if version == "" {
		version = "v3.1.0"
	}

	toolsDir := filepath.Join("build", "_output", "tools", "editorconfig-checker-"+version)
	checkerPath := filepath.Join(toolsDir, "editorconfig-checker")

	if _, err := os.Stat(checkerPath); os.IsNotExist(err) {
		a.Log("Downloading editorconfig-checker", version)
		args := download.Args{
			Args: install.Args{
				Asset: github.Asset{
					Release: github.Release{
						Tag: version,
						Repository: github.Repository{
							Owner: "editorconfig-checker",
							Repo:  "editorconfig-checker",
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

	cmd := exec.CommandContext(a.Context(), checkerPath)
	cmd.Stdout = a.Output()
	cmd.Stderr = a.Output()
	if err := cmd.Run(); err != nil {
		a.Fatal(err)
	}
}

func runGolangciLint(a *goyek.A) {
	version := os.Getenv("GOLANGCI_LINT_VERSION")
	if version == "" {
		version = "v2.4.0"
	}

	toolsDir := filepath.Join("build", "_output", "tools", "golangci-lint-"+version)
	linterPath := filepath.Join(toolsDir, "golangci-lint")

	if _, err := os.Stat(linterPath); os.IsNotExist(err) {
		a.Log("Downloading golangci-lint", version)
		args := download.Args{
			Args: install.Args{
				Asset: github.Asset{
					Release: github.Release{
						Tag: version,
						Repository: github.Repository{
							Owner: "golangci",
							Repo:  "golangci-lint",
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

	cmd := exec.CommandContext(a.Context(), linterPath, "run", "./...")
	cmd.Stdout = a.Output()
	cmd.Stderr = a.Output()
	if err := cmd.Run(); err != nil {
		a.Fatal(err)
	}
}
