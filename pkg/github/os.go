package github

import (
	"runtime"

	"github.com/cardil/ghet/pkg/match"
)

type OsFamily string

const (
	OSFamilyDarwin  OsFamily = "darwin"
	OSFamilyLinux   OsFamily = "linux"
	OSFamilyWindows OsFamily = "windows"
)

type OperatingSystem string

const (
	OSDarwin    OperatingSystem = "darwin"
	OSLinuxMusl OperatingSystem = "linux-musl"
	OSLinuxGnu  OperatingSystem = "linux-gnu"
	OSWindows   OperatingSystem = "windows"
)

func (os OperatingSystem) Matches(name string) bool {
	return matchWith(name, osMatchers[os])
}

func noOsMatches(name string) bool {
	oss := []OperatingSystem{
		OSDarwin,
		OSLinuxMusl,
		OSLinuxGnu,
		OSWindows,
	}
	for _, os := range oss {
		if os.Matches(name) {
			return false
		}
	}
	return true
}

func CurrentOS() OperatingSystem {
	family := OsFamily(runtime.GOOS)
	//goland:noinspection GoBoolExpressions
	if family == OSFamilyLinux {
		return linuxFlavor()
	}
	return OperatingSystem(family)
}

var osMatchers = map[OperatingSystem]match.Matcher{ //nolint:gochecknoglobals
	OSLinuxMusl: match.Every(
		match.Any(match.Substr("linux", "musl")),
		notPackageManagers,
	),
	OSLinuxGnu: match.Every(
		match.Any(
			match.Substr("linux", "glibc"),
			match.Substr("linux", "gnu"),
			match.Every(
				match.Substr("linux"),
				match.Not(match.Substr("musl")),
			),
		),
		notPackageManagers,
	),
	OSDarwin: match.Any(
		match.Substr("darwin"),
		match.Substr("mac"),
		match.Substr("osx"),
	),
	OSWindows: match.Any(
		match.Substr("win"),
	),
}
