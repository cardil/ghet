//go:build linux

package github

import (
	"strings"

	"github.com/u-root/u-root/pkg/ldd"
)

func linuxFlavor() OperatingSystem {
	fis, err := ldd.FList("/bin/sh")
	if err == nil {
		for _, fi := range fis {
			if strings.Contains(fi, "musl") {
				return OSLinuxMusl
			}
		}
	}

	return OSLinuxGnu
}
