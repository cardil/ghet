package dir

import "context"

func WithConfigDir(ctx context.Context, p string) context.Context {
	return context.WithValue(ctx, configDirKey, p)
}

func WithCacheDir(ctx context.Context, p string) context.Context {
	return context.WithValue(ctx, cacheDirKey, p)
}
