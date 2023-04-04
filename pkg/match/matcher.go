package match

type MatcherFn func(name string) bool

func (m MatcherFn) Matches(name string) bool {
	return m(name)
}

type Matcher interface {
	Matches(name string) bool
}

func Negate(matcher Matcher) Matcher {
	return MatcherFn(func(name string) bool {
		return !matcher.Matches(name)
	})
}

func Any(matchers ...Matcher) Matcher {
	return MatcherFn(func(name string) bool {
		for _, matcher := range matchers {
			if matcher.Matches(name) {
				return true
			}
		}
		return false
	})
}

func Every(matchers ...Matcher) Matcher {
	return MatcherFn(func(name string) bool {
		m := true
		for _, matcher := range matchers {
			m = m && matcher.Matches(name)
		}
		return m
	})
}
