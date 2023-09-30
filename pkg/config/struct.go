package config

type Config struct {
	Sites []Site `json:"sites"`
}

func (c Config) Site(site string) Site {
	for _, s := range c.Sites {
		if s.Address == site {
			return s
		}
	}
	return Site{}
}

type Type string

const (
	TypeGitHub Type = "github"
)

type Site struct {
	Type    `json:"type"`
	Address string `json:"address"`
	*Auth   `json:"auth"`
}

type Auth struct {
	Token string `json:"token"`
}

func (a *Auth) EffectiveToken() string {
	if a == nil {
		return ""
	}
	return a.Token
}

func (a *Auth) copy() *Auth {
	i := Auth{}
	if a.Token != "" {
		i.Token = a.Token
	}
	return &i
}
