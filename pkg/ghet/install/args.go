package install

import (
	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/github"
)

type Args struct {
	github.Asset
	config.Site
}
