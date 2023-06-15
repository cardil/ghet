package output

import (
	"context"
	"os"
	"path"

	configdir "github.com/cardil/ghet/pkg/config/dir"
)

type logFileKey struct{}

func EnsureLogFile(ctx context.Context) context.Context {
	if f := LogFileFrom(ctx); f == nil {
		f = createLogFile(ctx)
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

func createLogFile(ctx context.Context) *os.File {
	cachePath := configdir.Cache(ctx)
	logPath := path.Join(cachePath, "last-exec.log.jsonl")
	if logFile, err := os.Create(logPath); err != nil {
		panic(err)
	} else {
		return logFile
	}
}
