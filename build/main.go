package main

import (
	"os"
	"path/filepath"
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
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to determine source file path")
	}
	return filepath.Dir(filepath.Dir(file))
}
