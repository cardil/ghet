package context

import (
	sysctx "context"
)

// Context is a wrapper around the standard context.Context interface.
type Context interface {
	sysctx.Context
}

// WithValue returns a copy of parent in which the value associated with key is val.
func WithValue(parent Context, key, val any) Context {
	return sysctx.WithValue(parent, key, val)
}
