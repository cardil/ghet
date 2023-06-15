package api

import (
	"github.com/cardil/ghet/pkg/match"
)

type Asset struct {
	ID          int64
	Name        string
	ContentType string
	Size        int
	URL         string
}

func (a Asset) String() string {
	return a.Name
}

type IndexedAssets struct {
	Archives  []Asset
	Checksums []Asset
	Binaries  []Asset
}

func CreateIndex(assets []Asset) IndexedAssets {
	index := IndexedAssets{}
	for _, asset := range assets {
		name := asset.Name
		switch {
		case isArchive().Matches(name):
			index.Archives = append(index.Archives, asset)
		case isChecksum().Matches(name):
			index.Checksums = append(index.Checksums, asset)
		default:
			index.Binaries = append(index.Binaries, asset)
		}
	}
	return index
}

func isArchive() match.Matcher {
	return match.Any(
		match.EndsWith(".gz"),
		match.EndsWith(".tgz"),
		match.EndsWith(".bz2"),
		match.EndsWith(".tbz2"),
		match.EndsWith(".xz"),
		match.EndsWith(".txz"),
		match.EndsWith(".lz"),
		match.EndsWith(".tlz"),
		match.EndsWith(".lz4"),
		match.EndsWith(".zip"),
		match.EndsWith(".7z"),
		match.EndsWith(".rar"),
	)
}

func isChecksum() match.Matcher {
	return match.Any(
		match.EndsWith(".sha256"),
		match.EndsWith(".sha512"),
		match.EndsWith(".md5"),
		match.Regex("checksums?\\.txt"),
	)
}
