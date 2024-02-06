//go:build !linux

package github

func linuxFlavor() OperatingSystem {
	return OSLinuxGnu
}
