//go:build linux

package github

import (
	"strings"

	"github.com/u-root/u-root/pkg/ldd"
)

func linuxFlavor() OperatingSystem {
	if fis, err := ldd.FList("/bin/sh"); err == nil {
		for _, fi := range fis {
			if strings.Contains(fi, "musl") {
				return OSLinuxMusl
			}
		}
	}
	return OSLinuxGnu
}
