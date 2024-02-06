package config

import (
	"context"
	"fmt"
	"os"

	"emperror.dev/errors"
	"knative.dev/client-pkg/pkg/output/logging"
	"sigs.k8s.io/yaml"
)

// ErrInvalidConfigFile is returned when the config file is invalid.
var ErrInvalidConfigFile = errors.New("invalid config file")

func Load(ctx context.Context, file string) (Config, error) {
	l := logging.LoggerFrom(ctx).
		WithFields(logging.Fields{"configPath": file})
	l.Debug("Loading config as YAML")
	defaults := Config{
		Sites: []Site{{
			Type:    TypeGitHub,
			Address: "github.com",
		}},
	}
	var cfg Config
	if fileNotExists(file) {
		l.Debug("Config file does not exist, using defaults")
		return defaults, nil
	}
	bytes, err := os.ReadFile(file)
	if err != nil {
		return Config{}, asInvalidConfigErr(err)
	}
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return Config{}, asInvalidConfigErr(err)
	}

	return defaults.Merge(cfg), nil
}

func fileNotExists(file string) bool {
	_, err := os.Stat(file)
	return err != nil && os.IsNotExist(err)
}

func asInvalidConfigErr(err error) error {
	if errors.Is(err, ErrInvalidConfigFile) {
		return err
	}
	return errors.WithStack(
		errors.Wrap(ErrInvalidConfigFile, fmt.Sprintf("%+v", err)),
	)
}
