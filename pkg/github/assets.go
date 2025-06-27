package github

import (
	"path"
	"strings"

	"github.com/cardil/ghet/pkg/match"
)

type FileName struct {
	BaseName  string
	Extension string
}

func NewFileName(s string) FileName {
	basename := s
	ext := path.Ext(s)
	if ext != "" {
		basename = strings.TrimSuffix(s, ext)
	}
	return FileName{
		BaseName:  basename,
		Extension: ext,
	}
}

func (n FileName) ToString() string {
	if n.Extension == "" {
		return n.BaseName
	}
	joiner := "."
	if strings.HasPrefix(n.Extension, ".") {
		joiner = ""
	}
	return n.BaseName + joiner + n.Extension
}

func (n FileName) isEmpty() bool {
	return n.BaseName == "" && n.Extension == ""
}

type Checksums struct {
	FileName
}

func emptyChecksums() []Checksums {
	return []Checksums{
		{NewFileName("checksums.txt")},
		{NewFileName("checksums.out")},
		{NewFileName("sha256sum.txt")},
		{NewFileName("sha256sum.out")},
		{NewFileName("sha512sum.txt")},
		{NewFileName("sha512sum.out")},
		{FileName{Extension: "sha256"}},
		{FileName{Extension: "sha256sum"}},
		{FileName{Extension: "sha512"}},
		{FileName{Extension: "sha512sum"}},
	}
}

func (c Checksums) matcher(basename string, arch Architecture, sys OperatingSystem) match.Matcher {
	if c.isEmpty() {
		cc := emptyChecksums()
		mm := make([]match.Matcher, len(cc))
		for i, ch := range cc {
			mm[i] = ch.matcher(basename, arch, sys)
		}
		return match.Any(mm...)
	}
	return match.MatcherFn(func(name string) bool {
		return c.ToString() == name ||
			strings.HasPrefix(name, basename) &&
				strings.HasSuffix(name, c.ToString()) &&
				((arch.Matches(name) && sys.Matches(name)) ||
					(noArchMatches(name) && noOsMatches(name)))
	})
}

type Asset struct {
	FileName
	Architecture
	OperatingSystem
	Release
	Checksums
}

func (a Asset) Matches(filename string) bool {
	name := strings.ToLower(filename)
	basename := strings.ToLower(a.BaseName)
	coords := strings.Trim(
		strings.TrimSuffix(
			strings.TrimPrefix(name, basename),
			basename,
		), "-_",
	)
	cm := a.matcher(basename, a.Architecture, a.OperatingSystem)
	return cm.Matches(name) || ((strings.HasPrefix(name, basename) ||
		strings.HasSuffix(name, basename)) &&
		a.Architecture.Matches(coords) &&
		a.OperatingSystem.Matches(coords))
}
