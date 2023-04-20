package download_test

import (
	"embed"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"testing"

	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/context"
	"github.com/cardil/ghet/pkg/ghet/download"
	"github.com/cardil/ghet/pkg/ghet/install"
	"github.com/cardil/ghet/pkg/github"
	ghapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/output"
	gh "github.com/google/go-github/v48/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	artifactRe = regexp.MustCompile("^([a-z0-9-]+)/([a-z0-9-]+)(?:@([a-z0-9._-]+))?(?:!!([a-z0-9._-]+))?$")
	downloadRe = regexp.MustCompile("^https://github.com/([a-z0-9-]+)/([a-z0-9-]+)/releases/download/([a-z0-9._-]+)/(.+)$")
)

//go:embed testdata/*
var fs embed.FS

func TestCreatePlan(t *testing.T) {
	t.Parallel()
	testCases := []createPlanTestCase{{
		name: "knative-sandbox/kn-plugin-event!!kn-event",
		want: result{version: "knative-v1.9.1", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name:        "checksums.txt",
				ContentType: "text/plain",
				Size:        615,
			}, {
				Name: "kn-event-darwin-arm64",
				Size: 58_727_008,
			}},
		}},
	}, {
		name: "knative-sandbox/kn-plugin-event@knative-v1.8.0!!kn-event",
		want: result{version: "knative-v1.8.0", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name: "checksums.txt",
				Size: 615,
			}, {
				Name: "kn-event-darwin-arm64",
				Size: 58_542_306,
			}},
		}},
	}, {
		name: "derailed/k9s",
		want: result{version: "v0.27.3", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name:        "checksums.txt",
				Size:        896,
				ContentType: "text/plain; charset=utf-8",
			}, {
				Name:        "k9s_Darwin_arm64.tar.gz",
				Size:        18_197_297,
				ContentType: "application/gzip",
			}},
		}},
	}, {
		name: "kubernetes/minikube",
		want: result{version: "v1.30.1", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name: "minikube-darwin-arm64.sha256",
				Size: 65,
			}, {
				Name:        "minikube-darwin-arm64.tar.gz",
				Size:        33_490_698,
				ContentType: "application/octet-stream",
			}},
		}},
	}, {
		name: "lsd-rs/lsd",
		arch: github.ArchAMD64,
		os:   github.OSLinuxMusl,
		want: result{version: "0.23.1", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name:        "lsd-0.23.1-x86_64-unknown-linux-musl.tar.gz",
				Size:        941_846,
				ContentType: "application/gzip",
			}},
		}},
	}, {
		name: "marwanhawari/ppath",
		want: result{version: "v0.0.3", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name:        "checksums.txt",
				Size:        394,
				ContentType: "text/plain; charset=utf-8",
			}, {
				Name:        "ppath-v0.0.3-darwin-arm64.tar.gz",
				Size:        625_832,
				ContentType: "application/gzip",
			}},
		}},
	}, {
		name: "sharkdp/diskus",
		arch: github.ArchAMD64,
		os:   github.OSDarwin,
		want: result{version: "v0.7.0", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name:        "diskus-v0.7.0-x86_64-apple-darwin.tar.gz",
				Size:        364_147,
				ContentType: "application/gzip",
			}},
		}},
	}, {
		name: "sharkdp/pastel",
		arch: github.ArchX86,
		os:   github.OSLinuxMusl,
		want: result{version: "v0.9.0", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name:        "pastel-v0.9.0-i686-unknown-linux-musl.tar.gz",
				Size:        621_135,
				ContentType: "application/gzip",
			}},
		}},
	}, {
		name: "asciinema/agg",
		arch: github.ArchAMD64,
		os:   github.OSDarwin,
		want: result{version: "v1.4.0", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name:        "agg-x86_64-apple-darwin",
				Size:        7_834_192,
				ContentType: "binary/octet-stream",
			}},
		}},
	}, {
		name: "golangci/golangci-lint",
		want: result{version: "v1.52.2", Plan: download.Plan{
			Assets: []ghapi.Asset{{
				Name:        "golangci-lint-1.52.2-darwin-arm64.tar.gz",
				Size:        9_866_130,
				ContentType: "application/gzip",
			}, {
				Name:        "golangci-lint-1.52.2-checksums.txt",
				Size:        5_150,
				ContentType: "text/plain; charset=utf-8",
			}},
		}},
	}}
	for _, tc := range testCases {
		t.Run(tc.name, tc.performTest())
	}
}

type result struct {
	version string
	download.Plan
}

func (tc createPlanTestCase) args(t testingT) createPlanArgs {
	m := artifactRe.FindStringSubmatch(tc.name)
	require.NotNilf(t, m, "invalid test name %q", tc.name)
	a := createPlanArgs{
		owner:    m[1],
		repo:     m[2],
		basename: m[4],
		tag:      m[3],
	}
	if a.basename == "" {
		a.basename = a.repo
	}
	if a.tag == "" {
		a.tag = github.LatestTag
	}
	return a
}

func (tc createPlanTestCase) performTest() func(*testing.T) {
	ctx := output.WithContext(context.TODO(), output.NewTestPrinter())
	return func(t *testing.T) {
		ctx = context.WithTestingT(ctx, t)
		t.Parallel()
		tc.resolve(t).performTest(ctx, t)
	}
}

func (tc createPlanTestCase) resolve(t testingT) resolvedCreatePlanTestCase {
	if tc.responses == nil {
		tc.responses = autoResponses
	}
	args := tc.args(t)
	return resolvedCreatePlanTestCase{
		args:      args,
		arch:      tc.arch,
		os:        tc.os,
		want:      tc.want,
		wantErr:   tc.wantErr,
		responses: tc.responses(t, args),
	}
}

func (tc resolvedCreatePlanTestCase) performTest(ctx context.Context, t testingT) {
	ghapi.WithTestClient(t, func(client *gh.Client, mux *http.ServeMux) {
		ctx = ghapi.WithContext(ctx, client)
		tc.configureHTTP(t, mux)
		if tc.arch == "" {
			tc.arch = github.ArchARM64
		}
		if tc.os == "" {
			tc.os = github.OSDarwin
		}
		args := download.Args{
			Args: install.Args{
				Asset: github.Asset{
					FileName:        github.FileName{BaseName: tc.args.basename},
					Architecture:    tc.arch,
					OperatingSystem: tc.os,
					Release: github.Release{
						Tag: tc.args.tag,
						Repository: github.Repository{
							Owner: tc.args.owner,
							Repo:  tc.args.repo,
						},
					},
				},
				Site: config.Site{Type: config.TypeGitHub},
			},
			Destination: t.TempDir(),
		}
		p, err := download.CreatePlan(ctx, args)
		assert.ErrorIs(t, err, tc.wantErr, "%+v", err)
		assert.EqualValues(t,
			tc.normalize(tc.want),
			tc.normalize(tc.simplify(t, *p)),
		)
	})
}

type testingT interface {
	TempDir() string
	require.TestingT
	context.TestingT
}

type createPlanArgs struct {
	owner    string
	repo     string
	tag      string
	basename string
}

type createPlanTestCase struct {
	name    string
	arch    github.Architecture
	os      github.OperatingSystem
	want    result
	wantErr error
	responses
}

type resolvedCreatePlanTestCase struct {
	args      createPlanArgs
	arch      github.Architecture
	os        github.OperatingSystem
	want      result
	wantErr   error
	responses []response
}

func get(uri, body string) response {
	return respond(http.MethodGet, uri, body)
}

func respond(method, uri, body string, headers ...header) response {
	return response{
		statusCode: 200,
		request: request{
			method:  method,
			uri:     uri,
			headers: headers,
		},
		contentType: "application/json",
		body:        body,
	}
}

type response struct {
	statusCode  int
	body        string
	contentType string
	request
}

type request struct {
	method  string
	uri     string
	headers []header
}

type header struct {
	name  string
	value string
}

func (tc resolvedCreatePlanTestCase) configureHTTP(t testingT, mux *http.ServeMux) {
	for _, resp := range tc.responses {
		mux.HandleFunc(resp.uri, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != resp.method {
				t.Errorf("method mismatch: %s != %s, request: %+v", resp.method, r.Method, r)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			for _, h := range resp.headers {
				if r.Header.Get(h.name) != h.value {
					t.Errorf("header expected: %s=%s, request: %+v", h.name, h.value, r)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			w.Header().Set("Content-Type", resp.contentType)
			_, _ = w.Write([]byte(resp.body))
		})
	}
}

func (tc resolvedCreatePlanTestCase) simplify(t testingT, dp download.Plan) result {
	ver := ""
	if len(dp.Assets) > 0 {
		m := downloadRe.FindStringSubmatch(dp.Assets[0].URL)
		require.NotNilf(t, m, "invalid download url %q", dp.Assets[0].URL)
		ver = m[3]
	}
	return result{
		version: ver,
		Plan:    dp,
	}
}

func (tc resolvedCreatePlanTestCase) normalize(r result) result {
	return normalize(r, []normalizer{
		noIds,
		noDownloadURLs,
		defaultContentType,
		sorted,
	})
}

type normalizer func(res result) result

func defaultContentType(r result) result {
	for i := range r.Assets {
		if r.Assets[i].ContentType == "" {
			r.Assets[i].ContentType = "application/octet-stream"
		}
	}
	return r
}

func sorted(r result) result {
	sort.Slice(r.Assets, func(i, j int) bool {
		return r.Assets[i].Name < r.Assets[j].Name
	})
	return r
}

func noIds(r result) result {
	for i := range r.Assets {
		r.Assets[i].ID = 0
	}
	return r
}

func noDownloadURLs(r result) result {
	for i := range r.Assets {
		r.Assets[i].URL = ""
	}
	return r
}

func normalize(p result, normalizers []normalizer) result {
	for _, fn := range normalizers {
		p = fn(p)
	}
	return p
}

func readTestfile(t testingT, name string) string {
	bytes, err := fs.ReadFile("testdata/" + name)
	require.NoError(t, err)
	return string(bytes)
}

type responses func(t testingT, args createPlanArgs) []response

func autoResponses(t testingT, args createPlanArgs) []response {
	reqPath := fmt.Sprintf("/repos/%s/%s/releases/tags/%s",
		args.owner, args.repo, args.tag)
	if args.tag == github.LatestTag {
		reqPath = fmt.Sprintf("/repos/%s/%s/releases/latest",
			args.owner, args.repo)
	}
	testfile := fmt.Sprintf("GET-%s-%s-releases-%s.json",
		args.owner, args.repo, args.tag)
	return []response{
		get(reqPath, readTestfile(t, testfile)),
	}
}
