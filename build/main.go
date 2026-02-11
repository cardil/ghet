package main

import (
	"os"
	"path"
	"runtime"

	"github.com/cardil/ghet/build/pipelines"
	"github.com/goyek/goyek/v2"
)

func main() {
	if err := os.Chdir(rootDir()); err != nil {
		panic(err)
	}
	goyek.DefaultFlow = pipelines.Default()
	goyek.Main(os.Args[1:])
}

func rootDir() string {
	_, file, _, _ := runtime.Caller(0) //nolint:dogsled
	return path.Dir(path.Dir(file))
}
