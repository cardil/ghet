package github

import "runtime"

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
	Arch386     Architecture = "386"
	ArchAMD64   Architecture = "amd64"
	ArchARM     Architecture = "arm"
	ArchARM64   Architecture = "arm64"
	ArchPPC64LE Architecture = "ppc64le"
	ArchS390X   Architecture = "s390x"
)

func CurrentArchitecture() Architecture {
	return Architecture(runtime.GOARCH)
}

type OperatingSystem string

const (
	OSDarwin  OperatingSystem = "darwin"
	OSLinux   OperatingSystem = "linux"
	OSWindows OperatingSystem = "windows"
)

func CurrentOS() OperatingSystem {
	return OperatingSystem(runtime.GOOS)
}
