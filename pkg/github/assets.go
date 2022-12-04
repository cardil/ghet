package github

type FileName struct {
	BaseName  string
	Extension string
}

type Checksums struct {
	FileName
	Release
}

type Asset struct {
	FileName
	Architecture
	OperatingSystem
	Release
	Checksums
}
