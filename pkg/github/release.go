package github

import (
	"runtime"
	"strings"

	"github.com/cardil/ghet/pkg/match"
	"github.com/u-root/u-root/pkg/ldd"
)

const LatestTag = "latest"

type Repository struct {
	Owner string
	Repo  string
}

type Release struct {
	Tag string
	Repository
}

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

var archMatchers = map[Architecture]match.Matcher{ //nolint:gochecknoglobals
	ArchX86: match.Any(
		match.Substr("x86"),
		match.Regex("i?[3-6]86"),
	),
	ArchAMD64: match.Any(
		match.Substr("amd64"),
		match.Substr("x86_64"),
	),
	ArchARM: match.Any(
		match.Substr("arm32"),
		match.Substr("arm"),
	),
	ArchARM64:   match.Any(match.Substr("arm64")),
	ArchPPC64LE: match.Any(match.Regex("ppc-?64-?(?:le)?")),
	ArchS390X:   match.Any(match.Substr("s390x")),
}

var notPackageManagers = match.Every( //nolint:gochecknoglobals
	match.Negate(match.EndsWith(".deb")),
	match.Negate(match.EndsWith(".rpm")),
)
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
				match.Negate(match.Substr("musl")),
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

func matchWith(name string, matcher match.Matcher) bool {
	name = strings.ToLower(name)
	return matcher.Matches(name)
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
