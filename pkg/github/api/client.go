package api

import (
	"context"
	"net/http"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

type clientKey struct{}

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

func FromContext(ctx context.Context) *github.Client {
	if cl, ok := ctx.Value(clientKey{}).(*github.Client); ok {
		return cl
	}
	return NewClient(ctx, "")
}

func WithContext(ctx context.Context, cl *github.Client) context.Context {
	return context.WithValue(ctx, clientKey{}, cl)
}
