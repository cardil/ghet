package github

import (
	"path"
	"strings"
)

type FileName struct {
	BaseName  string
	Extension string
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

type Checksums struct {
	FileName
}

type Asset struct {
	FileName
	Architecture
	OperatingSystem
	Release
	Checksums
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
