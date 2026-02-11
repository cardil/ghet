package tasks

import (
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

func Test() goyek.Task {
	return goyek.Task{
		Name:  "test",
		Usage: "Run tests",
		Action: func(a *goyek.A) {
			ver := "v1.13.0"
			if envver := os.Getenv("GOTESTSUM_VERSION"); envver != "" {
				ver = envver
			}
			tool := fmt.Sprintf("go run gotest.tools/gotestsum@%s", ver)
			cmd.Exec(a, fmt.Sprint(tool, " --format testname ./..."))
		},
	}
}
