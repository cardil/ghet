package output

import (
	"context"
	"os"

	slog "github.com/go-eden/slf4go"
	sl "github.com/go-eden/slf4go-logrus"
	"github.com/sirupsen/logrus"
)

type loggerKey struct{}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	if !ok {
		prtr := FromContext(ctx)
		setupLogging(prtr)
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

func setupLogging(outs StandardOutputs) {
	sl.Init()
	logrus.SetOutput(outs.ErrOrStderr())
	l := logrus.WarnLevel
	var err error
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		l, err = logrus.ParseLevel(lvl)
		if err != nil {
			logrus.WithError(err).Error("Failed to parse LOG_LEVEL")
		}
	}
	logrus.SetLevel(l)
}
