package output

import (
	"context"

	slog "github.com/go-eden/slf4go"
)

type loggerKey struct{}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	if !ok {
		return slog.GetLogger()
	}
	return l
}

func EnsureLogger(ctx context.Context, fields ...slog.Fields) context.Context {
	l := LoggerFrom(ctx)
	for _, f := range fields {
		l = l.WithFields(f)
	}
	return WithLogger(ctx, l)
}
