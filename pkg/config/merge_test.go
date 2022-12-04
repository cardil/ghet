package config_test

import (
	"testing"

	"github.com/cardil/ghet/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	const (
		token     = "token"
		ghaddress = "github.com"
	)
	cfg := config.Config{
		Sites: []config.Site{{
			Type: config.TypeGitHub,
			Auth: &config.Auth{
				Token: token,
			},
		}},
	}
	defaults := config.Config{
		Sites: []config.Site{{
			Type:    config.TypeGitHub,
			Address: ghaddress,
		}},
	}
	merged := defaults.Merge(cfg)

	assert.Len(t, merged.Sites, 1)
	assert.Equal(t, config.Site{
		Type:    config.TypeGitHub,
		Address: ghaddress,
		Auth: &config.Auth{
			Token: token,
		},
	}, merged.Sites[0])
}
