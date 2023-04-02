package output

import (
	"context"
	"os"
	"path"

	"github.com/cardil/ghet/pkg/metadata"
	"github.com/kirsle/configdir"
)

type logFileKey struct{}

func EnsureLogFile(ctx context.Context) context.Context {
	if f := LogFileFrom(ctx); f == nil {
		f = createLogFile()
		return WithLogFile(ctx, f)
	}
	return ctx
}

func LogFileFrom(ctx context.Context) *os.File {
	if f, ok := ctx.Value(logFileKey{}).(*os.File); ok {
		return f
	}
	return nil
}

func WithLogFile(ctx context.Context, f *os.File) context.Context {
	return context.WithValue(ctx, logFileKey{}, f)
}

func createLogFile() *os.File {
	cachePath := configdir.LocalCache(metadata.Name)
	// Ensure it exists.
	if err := configdir.MakePath(cachePath); err != nil {
		panic(err)
	}
	logPath := path.Join(cachePath, "last-exec.log.jsonl")
	if logFile, err := os.Create(logPath); err != nil {
		panic(err)
	} else {
		return logFile
	}
}
