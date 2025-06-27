package install_test

import (
	"testing"

	"github.com/cardil/ghet/pkg/ghet/install"
	"github.com/cardil/ghet/pkg/github"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []testcase{{
		args: "cardil/ghet",
		want: install.Installation{
			Asset: github.Asset{
				FileName: github.FileName{
					BaseName:  "ghet",
					Extension: "",
				},
				Release: github.Release{
					Tag: "latest",
					Repository: github.Repository{
						Owner: "cardil",
						Repo:  "ghet",
					},
				},
			},
		},
	}, {
		args: "owner/repo@version::archive-name!!binary-name",
		want: install.Installation{
			Asset: github.Asset{
				FileName: github.FileName{
					BaseName:  "binary-name",
					Extension: "",
				},
				Release: github.Release{
					Tag: "version",
					Repository: github.Repository{
						Owner: "owner",
						Repo:  "repo",
					},
				},
			},
		},
	}, {
		args: "owner/repo@version",
		want: install.Installation{
			Asset: github.Asset{
				FileName: github.FileName{
					BaseName: "repo",
				},
				Release: github.Release{
					Tag: "version",
					Repository: github.Repository{
						Owner: "owner",
						Repo:  "repo",
					},
				},
			},
		},
	}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.args, func(t *testing.T) {
			t.Parallel()
			got := install.Parse(tt.args)
			want := tt.want.WithDefaults().Asset
			if got.Asset != want {
				t.Errorf("Parse()\n  got = %#v,\n want = %#v", got.Asset, want)
			}
		})
	}
}

type testcase struct {
	args string
	want install.Installation
}
