package match

import "strings"

func Substr(sub ...string) Matcher {
	mchs := make([]Matcher, len(sub))
	for i, s := range sub {
		substr := s
		mchs[i] = MatcherFn(func(name string) bool {
			return strings.Contains(name, substr)
		})
	}
	return Every(mchs...)
}

func EndsWith(suffix string) Matcher {
	return MatcherFn(func(name string) bool {
		return strings.HasSuffix(name, suffix)
	})
}
