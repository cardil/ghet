//go:build !race

package download_test

import (
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	configdir "github.com/cardil/ghet/pkg/config/dir"
	"github.com/cardil/ghet/pkg/ghet/download"
	"github.com/cardil/ghet/pkg/ghet/install"
	pkggithub "github.com/cardil/ghet/pkg/github"
	ghapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/google/go-github/v48/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"knative.dev/client/pkg/context"
	"knative.dev/client/pkg/output"
)

func TestDownload(t *testing.T) {
	t.Parallel()
	tcs := []downloadTestCase{{
		name: "kubernetes/minikube",
		args: downloadArgs{
			name: "minikube",
			assets: []string{
				"minikube-linux-amd64.tar.gz",
				"minikube-linux-amd64.sha256",
			},
			verifyInArchive: true,
		},
		want: []downloaded{{
			name: "minikube",
			size: 42,
		}},
	}, {
		name: "sharkdp/diskus",
		args: downloadArgs{
			name: "diskus",
			assets: []string{
				"diskus-v0.7.0-x86_64-unknown-linux-gnu.tar.gz",
			},
		},
		want: []downloaded{{
			name: "diskus",
			size: 40,
		}},
	}, {
		name: "pulumi/pulumi",
		args: downloadArgs{
			name: "pulumi",
			assets: []string{
				"pulumi-v3.71.0-linux-x64.tar.gz",
				"pulumi-3.71.0-checksums.txt",
			},
			multipleBins: true,
		},
		want: []downloaded{
			{name: "pulumi", size: 40},
			{name: "pulumi-watch", size: 49},
			{name: "pulumi-language-go", size: 54},
		},
	}, {
		name: "knative-sandbox/kn-plugin-event",
		args: downloadArgs{
			name: "kn-event",
			assets: []string{
				"kn-event-linux-amd64",
				"kn-event-checksums.txt",
			},
		},
		want: []downloaded{{
			name: "kn-event",
			size: 42,
		}},
	}, {
		name: "asciinema/agg",
		args: downloadArgs{
			name: "agg",
			assets: []string{
				"agg-x86_64-unknown-linux-gnu",
			},
		},
		want: []downloaded{{
			name: "agg",
			size: 37,
		}},
	}}
	for _, tc := range tcs {
		t.Run(tc.name, tc.run)
	}
}

func (tc downloadTestCase) run(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	ctx := context.TestContext(t)
	ctx = configdir.WithCacheDir(ctx, tmpDir)
	ctx = configdir.WithConfigDir(ctx, tmpDir)
	ctx = output.WithContext(ctx, output.NewTestPrinter())
	ghapi.WithTestClient(t, func(client *github.Client, mux *http.ServeMux) {
		ctx = ghapi.WithContext(ctx, client)
		tc.configureMux(mux)

		wd := t.TempDir()
		plan := tc.buildPlan(t, client.BaseURL)
		args := tc.buildArgs(wd)
		err := plan.Download(ctx, args)
		assert.ErrorIs(t, err, tc.wantErr, "%+v", err)
		for _, d := range tc.want {
			fp := path.Join(wd, d.name)
			var fi os.FileInfo
			fi, err = os.Stat(fp)
			require.NoError(t, err)
			assert.Equal(t, d.size, fi.Size(), "file %s has wrong size", d.name)
			assert.True(t, fi.Mode().IsRegular(), "file %s is not regular", d.name)
			assert.True(t, isExecutable(fi.Mode()), "file %s is not executable", d.name)
		}
	})
}

func (tc downloadTestCase) configureMux(mux *http.ServeMux) {
	for i := range tc.args.assets {
		asset := tc.args.assets[i]
		mux.HandleFunc("/"+asset, func(w http.ResponseWriter, r *http.Request) {
			f, err := fs.Open("testdata/" + asset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			defer f.Close()
			st, err := f.Stat()
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			http.ServeContent(w, r, asset, st.ModTime(), f.(io.ReadSeeker))
		})
	}
}

func (tc downloadTestCase) buildPlan(t testingT, baseurl *url.URL) download.Plan {
	p := download.Plan{
		Assets: make([]ghapi.Asset, 0, len(tc.args.assets)),
	}
	// Deterministic random numbers.
	rnd := rand.New(rand.NewSource(465712)) //nolint:gosec
	for _, asset := range tc.args.assets {
		f, err := fs.Open("testdata/" + asset)
		require.NoError(t, err)
		st, err := f.Stat()
		require.NoError(t, err)
		require.NoError(t, f.Close())

		p.Assets = append(p.Assets, ghapi.Asset{
			ID:          rnd.Int63(),
			Name:        asset,
			ContentType: "application/octet-stream",
			Size:        int(st.Size()),
			URL:         baseurl.String() + asset,
		})
	}
	return p
}

func (tc downloadTestCase) buildArgs(wd string) download.Args {
	fields := strings.FieldsFunc(tc.name, func(r rune) bool {
		return r == '/'
	})
	repo := pkggithub.Repository{
		Owner: fields[0],
		Repo:  fields[1],
	}
	return download.Args{
		Args: install.Args{
			Asset: pkggithub.Asset{
				FileName: pkggithub.FileName{
					BaseName: tc.args.name,
				},
				Release:         pkggithub.Release{Repository: repo},
				Architecture:    pkggithub.ArchAMD64,
				OperatingSystem: pkggithub.OSLinuxGnu,
			},
			MultipleBinaries: tc.args.multipleBins,
			VerifyInArchive:  tc.args.verifyInArchive,
		},
		Destination: wd,
	}
}

type downloadArgs struct {
	name            string
	assets          []string
	multipleBins    bool
	verifyInArchive bool
}

type downloaded struct {
	name string
	size int64
}

type downloadTestCase struct {
	name string
	args downloadArgs

	want    []downloaded
	wantErr error
}

func isExecutable(mode os.FileMode) bool {
	return mode&0o111 != 0
}
