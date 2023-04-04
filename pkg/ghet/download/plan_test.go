package download_test

import (
	"embed"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/context"
	"github.com/cardil/ghet/pkg/ghet/download"
	"github.com/cardil/ghet/pkg/ghet/install"
	pkggithub "github.com/cardil/ghet/pkg/github"
	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/google/go-github/v48/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/*
var fs embed.FS

func TestCreatePlan(t *testing.T) {
	t.Parallel()
	testCases := []createPlanTestCase{{
		name: "latest release of kn-event",
		want: download.Plan{
			Assets: []githubapi.Asset{{
				Name:        "checksums.txt",
				URL:         downloadURL("1.9.1", "checksums.txt"),
				ContentType: "text/plain",
				Size:        615,
			}, {
				Name:        "kn-event-darwin-arm64",
				URL:         downloadURL("1.9.1", "kn-event-darwin-arm64"),
				ContentType: "application/octet-stream",
				Size:        58727008,
			}},
		},
		responses: []response{
			get("/repos/knative-sandbox/kn-plugin-event/releases/latest",
				testfile(t, "GET-releases-latest.json")),
		},
	}, {
		name: "1.8 release of kn-event",
		args: func(args download.Args) download.Args {
			args.Tag = "knative-v1.8.0"
			return args
		},
		want: download.Plan{
			Assets: []githubapi.Asset{{
				Name:        "checksums.txt",
				URL:         downloadURL("1.8.0", "checksums.txt"),
				ContentType: "application/octet-stream",
				Size:        615,
			}, {
				Name:        "kn-event-darwin-arm64",
				URL:         downloadURL("1.8.0", "kn-event-darwin-arm64"),
				ContentType: "application/octet-stream",
				Size:        58_542_306,
			}},
		},
		responses: []response{
			get("/repos/knative-sandbox/kn-plugin-event/releases/tags/knative-v1.8.0",
				testfile(t, "GET-release-1.8.json")),
		},
	}}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			performCreatePlanTest(t, tc)
		})
	}
}

func performCreatePlanTest(t testingT, tc createPlanTestCase) {
	githubapi.WithTestClient(t, func(client *github.Client, mux *http.ServeMux) {
		ctx := githubapi.WithContext(context.TestContext(t), client)
		tc.configureHTTP(t, mux)
		defaultArgs := download.Args{
			Args: install.Args{
				Asset: pkggithub.Asset{
					FileName: pkggithub.FileName{
						BaseName: "kn-event",
					},
					Architecture:    pkggithub.ArchARM64,
					OperatingSystem: pkggithub.OSDarwin,
					Release: pkggithub.Release{
						Tag: pkggithub.LatestTag,
						Repository: pkggithub.Repository{
							Owner: "knative-sandbox",
							Repo:  "kn-plugin-event",
						},
					},
					Checksums: pkggithub.Checksums{
						FileName: pkggithub.FileName{
							BaseName:  "checksums",
							Extension: "txt",
						},
					},
				},
				Site: config.Site{
					Type: config.TypeGitHub,
				},
			},
			Destination: t.TempDir(),
		}
		plan, err := download.CreatePlan(ctx, tc.buildArgs(defaultArgs))
		assert.ErrorIs(t, err, tc.wantErr, "%+v", err)
		assert.EqualValues(t, tc.normalize(tc.want), tc.normalize(*plan))
	})
}

func downloadURL(ver, asset string) string {
	return fmt.Sprintf("https://github.com/"+
		"knative-sandbox/kn-plugin-event/releases/download/knative-v%s/%s", ver, asset)
}

type testingT interface {
	TempDir() string
	require.TestingT
	context.TestingT
}

type createPlanTestCase struct {
	name        string
	args        func(defaults download.Args) download.Args
	responses   []response
	want        download.Plan
	normalizers []normalizer
	wantErr     error
}

func get(uri, body string, headers ...header) response {
	return respond(http.MethodGet, uri, body, headers...)
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

func (tc createPlanTestCase) configureHTTP(t testingT, mux *http.ServeMux) {
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

func (tc createPlanTestCase) buildArgs(defaults download.Args) download.Args {
	if tc.args != nil {
		return tc.args(defaults)
	}
	return defaults
}

func (tc createPlanTestCase) normalize(plan download.Plan) download.Plan {
	n := tc.normalizers
	if len(tc.normalizers) == 0 {
		n = []normalizer{noIDs, sorted}
	}
	return normalize(plan, n)
}

type normalizer func(plan download.Plan) download.Plan

func noIDs(plan download.Plan) download.Plan {
	for i := range plan.Assets {
		plan.Assets[i].ID = 0
	}
	return plan
}

func sorted(plan download.Plan) download.Plan {
	sort.Slice(plan.Assets, func(i, j int) bool {
		return plan.Assets[i].Name < plan.Assets[j].Name
	})
	return plan
}

func normalize(plan download.Plan, normalizers []normalizer) download.Plan {
	for _, normlizr := range normalizers {
		plan = normlizr(plan)
	}
	// for i := range plan.Assets {
	// 	plan.Assets[i].ID = 0
	// }
	return plan
}

func testfile(t testingT, name string) string {
	bytes, err := fs.ReadFile("testdata/" + name)
	require.NoError(t, err)
	return string(bytes)
}
