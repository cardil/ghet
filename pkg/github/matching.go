package github

import (
	"strings"

	"github.com/cardil/ghet/pkg/match"
)

var notPackageManagers = match.Every( //nolint:gochecknoglobals
	match.Not(match.EndsWith(".deb")),
	match.Not(match.EndsWith(".rpm")),
)

func matchWith(name string, matcher match.Matcher) bool {
	name = strings.ToLower(name)
	return matcher.Matches(name)
}
