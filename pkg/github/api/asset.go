package api

import "github.com/cardil/ghet/pkg/match"

type Asset struct {
	ID          int64
	Name        string
	ContentType string
	Size        int
	URL         string
}

type IndexedAssets struct {
	Archives  []Asset
	Checksums []Asset
	Other     []Asset
}

func CreateIndex(assets []Asset) IndexedAssets {
	index := IndexedAssets{}
	for _, asset := range assets {
		name := asset.Name
		if isArchive.Matches(name) {
			index.Archives = append(index.Archives, asset)
		} else if isChecksum.Matches(name) {
			index.Checksums = append(index.Checksums, asset)
		} else {
			index.Other = append(index.Other, asset)
		}
	}
	return index
}

var (
	isArchive = match.Any(
		match.EndsWith(".tar.gz"),
		match.EndsWith(".tgz"),
		match.EndsWith(".tar.bz2"),
		match.EndsWith(".tbz2"),
		match.EndsWith(".tar.xz"),
		match.EndsWith(".txz"),
		match.EndsWith(".zip"),
	)

	isChecksum = match.Any(
		match.EndsWith(".sha256"),
		match.EndsWith(".sha512"),
		match.EndsWith(".md5"),
		match.Regex("checksums?\\.txt"),
	)
)
