package config

import "context"

var configKey = struct{}{}

func FromContext(ctx context.Context) Config {
	p, ok := ctx.Value(configKey).(Config)
	if !ok {
		return Config{}
	}
	return p
}

func WithContext(ctx context.Context, c Config) context.Context {
	return context.WithValue(ctx, configKey, c)
}
