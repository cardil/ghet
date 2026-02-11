package tasks

import (
	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

func Build() goyek.Task {
	return goyek.Task{
		Name:  "build",
		Usage: "Build the project",
		Action: func(a *goyek.A) {
			cmd.Exec(a, "go build -v ./...")
		},
	}
}
