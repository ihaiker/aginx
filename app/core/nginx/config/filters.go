package config

import "regexp"

type CharFilter func(current, previous string) bool

func (self CharFilter) And(cf ...CharFilter) CharFilter {
	return func(current, previous string) bool {
		if !self(current, previous) {
			return false
		}
		for _, filter := range cf {
			if !filter(current, previous) {
				return false
			}
		}
		return true
	}
}

func (self CharFilter) Or(cf ...CharFilter) CharFilter {
	return func(current, previous string) bool {
		out := self(current, previous)
		for _, filter := range cf {
			out = out || filter(current, previous)
		}
		return out
	}
}

var (
	vailCharRegexp            = regexp.MustCompile("\\S")
	ValidChars     CharFilter = func(current, previous string) bool {
		return vailCharRegexp.MatchString(current)
	}

	In = func(chars ...string) CharFilter {
		return func(current, previous string) bool {
			for _, char := range chars {
				if char == current {
					return true
				}
			}
			return false
		}
	}

	Not = func(cf CharFilter) CharFilter {
		return func(current, previous string) bool {
			return !cf(current, previous)
		}
	}
)
