package dir

import (
	"context"
	"log"
	"os"

	"emperror.dev/errors"
	"github.com/cardil/ghet/pkg/metadata"
	"github.com/kirsle/configdir"
)

const (
	ConfigDirEnvName = "GHET_CONFIG_DIR"
	CacheDirEnvName  = "GHET_CACHE_DIR"
)

type cacheDirKey struct{}

type configDirKey struct{}

func Config(ctx context.Context) string {
	return userPath(ctx, configDirKey{}, ConfigDirEnvName, func() string {
		return configdir.LocalConfig(metadata.Name)
	})
}

func Cache(ctx context.Context) string {
	return userPath(ctx, cacheDirKey{}, CacheDirEnvName, func() string {
		return configdir.LocalCache(metadata.Name)
	})
}

func userPath(ctx context.Context, key interface{}, envKey string, fn func() string) string {
	if p, ok := ctx.Value(key).(string); ok {
		return ensurePathExists(p)
	}
	p := os.Getenv(envKey)
	if p == "" {
		p = fn()
	}
	return ensurePathExists(p)
}

func ensurePathExists(p string) string {
	if err := configdir.MakePath(p); err != nil {
		log.Fatal(errors.WithStack(err))
	}
	return p
}
