package match

import (
	"regexp"
)

type regexMatcher struct {
	rxs []*regexp.Regexp
}

func (r regexMatcher) Matches(name string) bool {
	m := true
	for _, rx := range r.rxs {
		m = m && rx.MatchString(name)
	}
	return m
}

func Regex(regex ...string) Matcher {
	rxs := make([]*regexp.Regexp, len(regex))
	for i, r := range regex {
		rxs[i] = regexp.MustCompile(r)
	}
	return &regexMatcher{rxs}
}
