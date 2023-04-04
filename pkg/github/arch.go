package github

import (
	"runtime"

	"github.com/cardil/ghet/pkg/match"
)

type Architecture string

const (
	ArchX86     Architecture = "x86"
	ArchAMD64   Architecture = "amd64"
	ArchARM     Architecture = "arm"
	ArchARM64   Architecture = "arm64"
	ArchPPC64LE Architecture = "ppc64le"
	ArchS390X   Architecture = "s390x"
)

func (a Architecture) Matches(name string) bool {
	return matchWith(name, archMatchers[a])
}

func CurrentArchitecture() Architecture {
	return Architecture(runtime.GOARCH)
}

var archMatchers = map[Architecture]match.Matcher{ //nolint:gochecknoglobals
	ArchX86: match.Any(
		match.Every(
			match.Substr("x86"),
			match.Not(match.Substr("x86_64")),
		),
		match.Regex("i?[3-6]86"),
	),
	ArchAMD64: match.Any(
		match.Substr("amd64"),
		match.Substr("x86_64"),
	),
	ArchARM: match.Any(
		match.Substr("arm32"),
		match.Every(
			match.Substr("arm"),
			match.Not(match.Substr("arm64")),
		),
	),
	ArchARM64:   match.Any(match.Substr("arm64")),
	ArchPPC64LE: match.Any(match.Regex("ppc-?64-?(?:le)?")),
	ArchS390X:   match.Any(match.Substr("s390x")),
}
