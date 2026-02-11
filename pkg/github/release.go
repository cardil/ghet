package github

const LatestTag = "latest"

type Repository struct {
	Owner string
	Repo  string
}

type Release struct {
	Repository

	Tag string
}
