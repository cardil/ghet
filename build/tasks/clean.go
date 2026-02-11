package tasks

import (
	"os"

	"github.com/goyek/goyek/v2"
)

func Clean() goyek.Task {
	return goyek.Task{
		Name:  "clean",
		Usage: "Clean build artifacts",
		Action: func(a *goyek.A) {
			if err := os.RemoveAll("build/_output"); err != nil {
				a.Fatal(err)
			}
		},
	}
}
