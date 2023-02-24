package config

import (
	"context"
	"os"

	"github.com/cardil/ghet/pkg/output"
	log "github.com/go-eden/slf4go"
	"sigs.k8s.io/yaml"
)

func Load(ctx context.Context, file string) (Config, error) {
	l := output.LoggerFrom(ctx).
		WithFields(log.Fields{"configPath": file})
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
		return Config{}, err
	}
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return Config{}, err
	}

	return defaults.Merge(cfg), nil
}

func fileNotExists(file string) bool {
	_, err := os.Stat(file)
	return err != nil && os.IsNotExist(err)
}
