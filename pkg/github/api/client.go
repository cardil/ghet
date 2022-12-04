package api

import (
	"context"
	"net/http"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

func NewClient(ctx context.Context, token string) *github.Client {
	var httpClient *http.Client
	if token != "" {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(ctx, src)
	}

	return github.NewClient(httpClient)
}
