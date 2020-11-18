package util

import "strings"

func Split2(s, sep string) (a string, b string) {
	ary := strings.SplitN(s, sep, 2)
	a = ary[0]
	if len(ary) == 2 {
		b = ary[1]
	}
	return
}
func Split3(s, sep string) (a, b, c string) {
	ary := strings.SplitN(s, sep, 3)
	a = ary[0]
	if len(ary) > 1 {
		b = ary[1]
	}
	if len(ary) > 2 {
		c = ary[2]
	}
	return
}
