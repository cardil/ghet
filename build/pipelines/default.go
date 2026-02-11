package pipelines

import (
	"github.com/cardil/ghet/build/tasks"
	"github.com/goyek/goyek/v2"
)

func Default() *goyek.Flow {
	f := &goyek.Flow{}
	build := f.Define(tasks.Build())
	lint := f.Define(tasks.Lint())
	test := f.Define(tasks.Test())
	f.Define(tasks.Clean())

	// Default pipeline: build, lint, and test
	f.SetDefault(f.Define(goyek.Task{
		Name:  "all",
		Usage: "Build, lint, and test",
		Deps:  goyek.Deps{build, lint, test},
	}))
	return f
}
