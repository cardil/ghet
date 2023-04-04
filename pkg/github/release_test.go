package github_test

import (
	"testing"

	"github.com/cardil/ghet/pkg/github"
	"github.com/stretchr/testify/assert"
)

func TestOsLinuxMatch(t *testing.T) {
	cases := []testCaseOsLinuxMatch{{
		name: "pastel-v0.9.0-x86_64-pc-windows-gnu.zip",
	}, {
		name: "pastel-v0.9.0-x86_64-unknown-linux-gnu.tar.gz",
		want: true,
	}}
	os := github.OSLinuxGnu
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := os.Matches(tc.name)
			assert.Equal(t, tc.want, got)
		})
	}
}

type testCaseOsLinuxMatch struct {
	name string
	want bool
}
