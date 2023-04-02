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

func TestContext(t TestingT) sysctx.Context {
	return sysctx.WithValue(sysctx.TODO(), testingTKey{}, t)
}

func TestingTFromContext(ctx sysctx.Context) TestingT {
	t, ok := ctx.Value(testingTKey{}).(TestingT)
	if !ok {
		return nil
	}
	return t
}
