package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/google/go-github/v48/github"
	"github.com/stretchr/testify/require"
)

func WithTestClient(t require.TestingT, fn func(*github.Client, *http.ServeMux)) {
	client, mux, teardown := testClient(t)
	defer teardown()
	fn(client, mux)
}

func testClient(t require.TestingT) (*github.Client, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	client := github.NewClient(nil)
	u, err := url.Parse(server.URL + "/")
	if err != nil {
		require.NoError(t, err)
	}
	client.BaseURL = u
	client.UploadURL = u

	return client, mux, server.Close
}
