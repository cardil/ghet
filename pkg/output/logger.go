package output

import (
	"log"
	"os"

	"emperror.dev/errors"
	"github.com/cardil/ghet/pkg/context"
	slog "github.com/go-eden/slf4go"
	sz "github.com/go-eden/slf4go-zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

type loggerKey struct{}

func WithLogger(ctx context.Context, l slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func LoggerFrom(ctx context.Context) slog.Logger {
	l, ok := ctx.Value(loggerKey{}).(slog.Logger)
	if !ok {
		setupLogging(ctx)
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

func setupLogging(ctx context.Context) {
	var logger *zap.Logger
	if t := context.TestingTFromContext(ctx); t != nil {
		logger = createTestingLogger(t)
	} else {
		logger = teeLoggers(
			createDefaultLogger(ctx),
			createFileLogger(ctx),
		)
	}

	slog.SetDriver(&sz.ZapDriver{
		Logger: logger,
	})
}

func teeLoggers(logger1 *zap.Logger, logger2 *zap.Logger) *zap.Logger {
	return zap.New(zapcore.NewTee(
		logger1.Core(),
		logger2.Core(),
	))
}

func createFileLogger(ctx context.Context) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	logFile := LogFileFrom(ctx)
	if logFile == nil {
		panic(errors.New("no log file in context"))
	}
	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		zapcore.AddSync(logFile),
		zapcore.DebugLevel,
	))
}

func createDefaultLogger(ctx context.Context) *zap.Logger {
	prtr := PrinterFrom(ctx)
	errout := prtr.ErrOrStderr()
	ec := zap.NewDevelopmentEncoderConfig()
	ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	ec.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	ec.ConsoleSeparator = " "
	lvl := activeLogLevel(zapcore.WarnLevel)
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(ec),
		zapcore.AddSync(errout),
		lvl,
	))
	if !isTerminal(errout) {
		ec = zap.NewProductionEncoderConfig()
		logger = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(ec),
			zapcore.AddSync(errout),
			lvl,
		))
	}
	return logger
}

func createTestingLogger(t context.TestingT) *zap.Logger {
	lvl := activeLogLevel(zapcore.DebugLevel)
	return zaptest.NewLogger(t, zaptest.WrapOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(sz.SkipUntilTrueCaller),
	), zaptest.Level(lvl))
}

func activeLogLevel(defaultLevel zapcore.Level) zapcore.Level {
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		l, err := zapcore.ParseLevel(lvl)
		if err != nil {
			log.Fatal(errors.WithStack(err))
		}
		return l
	}
	return defaultLevel
}
