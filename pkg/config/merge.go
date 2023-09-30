package config

func (c Config) Merge(cfg Config) Config {
	if len(cfg.Sites) == 0 {
		return c
	}
	if len(c.Sites) == 0 {
		return cfg
	}
	matched := make([]pair, 0, len(cfg.Sites))
	unmatched := make([]Site, 0, len(cfg.Sites))
	for _, site := range c.Sites {
		found := false
		for _, cfgSite := range cfg.Sites {
			if cfgSite.Match(site) {
				matched = append(matched, pair{site, cfgSite})
				found = true
				break
			}
		}
		if !found {
			unmatched = append(unmatched, site)
		}
	}
	sites := make([]Site, 0, len(matched)+len(unmatched))
	for _, p := range matched {
		sites = append(sites, p.original.Merge(p.replacement))
	}
	sites = append(sites, unmatched...)
	return Config{Sites: sites}
}

func (s Site) Match(site Site) bool {
	if s.Address == "" {
		return s.Type == site.Type
	}
	return s.Type == site.Type && s.Address == site.Address
}

func (s Site) Merge(override Site) Site {
	if s.Type == "" {
		s.Type = override.Type
	}
	if s.Address == "" {
		s.Address = override.Address
	}
	if s.Auth == nil {
		s.Auth = override.Auth.copy()
	}
	return s
}

type pair struct {
	original, replacement Site
}
