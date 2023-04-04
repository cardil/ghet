package github

import (
	"runtime"
	"strings"

	"github.com/cardil/ghet/pkg/match"
	"github.com/u-root/u-root/pkg/ldd"
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

func linuxFlavor() OperatingSystem {
	if fis, err := ldd.Ldd([]string{"/bin/sh"}); err == nil {
		for _, fi := range fis {
			if strings.Contains(fi.Name(), "musl") {
				return OSLinuxMusl
			}
		}
	}
	return OSLinuxGnu
}
