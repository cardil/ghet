package context

import (
	sysctx "context"

	"go.uber.org/zap/zaptest"
)

type testingTKey struct{}

type TestingT interface {
	Log(args ...any)
	Cleanup(func())
	TempDir() string

	zaptest.TestingT
}

func TODO() sysctx.Context {
	return sysctx.TODO()
}

func TestContext(t TestingT) sysctx.Context {
	return WithTestingT(sysctx.TODO(), t)
}

func WithTestingT(ctx sysctx.Context, t TestingT) sysctx.Context {
	return sysctx.WithValue(ctx, testingTKey{}, t)
}

func TestingTFromContext(ctx sysctx.Context) TestingT {
	t, ok := ctx.Value(testingTKey{}).(TestingT)
	if !ok {
		return nil
	}
	return t
}
