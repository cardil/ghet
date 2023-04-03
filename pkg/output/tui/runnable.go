package tui

type Runnable[T any] interface {
	With(fn func(T) error) error
}
