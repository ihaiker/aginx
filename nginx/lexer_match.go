package nginx

import "strings"

func (a Args) Match(directive *Directive) bool {
	if !match(a.Arg.Comparison, directive.Args, a.Arg.Value) {
		return false
	}
	if a.Next != nil {
		ret := true
		for _, addition := range a.Next {
			if addition.Operator == "&" {
				ret = ret && match(addition.Arg.Comparison, directive.Args, addition.Arg.Value)
			} else if addition.Operator == "|" {
				ret = ret || match(addition.Arg.Comparison, directive.Args, addition.Arg.Value)
			} else { //fixme 这里用于扩展
				ret = false
			}
		}
		if ret == false {
			return false
		}
	}
	return true
}

func (c *Children) Match(directive *Directive) bool {
	if c.Directive != nil {
		return c.Directive.MatchAny(directive.Body)
	} else {
		ret := c.Group.First.MatchAny(directive.Body)
		for _, addition := range c.Group.Next {
			if addition.Operator == "&" {
				ret = ret && addition.Next.MatchAny(directive.Body)
			} else if addition.Operator == "|" {
				ret = ret || addition.Next.MatchAny(directive.Body)
			} else { //fixme 这里用于扩展
				ret = false
			}
		}

		if ret {
			return true
		}
	}
	return false
}

func (e *DirectiveExpr) Match(directive *Directive) bool {

	if e.Name != "" && !match(e.Comparison, []string{directive.Name}, e.Name) {
		return false
	} else if e.All == "*" {
		//true
	}

	if e.Args != nil {
		if !e.Args.Match(directive) {
			return false
		}
	}

	return true
}

func (e *DirectiveExpr) MatchAny(directive []*Directive) bool {
	for _, d := range directive {
		if e.Match(d) {
			return true
		}
	}
	return false
}

func (e *Expression) Match(directive *Directive) bool {
	if !e.Directive.Match(directive) {
		return false
	}

	if e.Children != nil {
		search := &Directive{Body: directive.Body}
		for _, child := range e.Children {
			match := &Directive{Body: make([]*Directive, 0)}
			if child.Match(search) {
				match.Body = append(match.Body, search.Body...)
			}
			if len(match.Body) > 0 {
				search = match
			} else {
				return false
			}
		}
	}

	return true
}

func match(comparison string, values []string, query string) bool {
	switch comparison {
	case "!":
		return !match("", values, query)
	case "@":
		for _, value := range values {
			if strings.Contains(value, query) {
				return true
			}
		}
	case "^":
		for _, value := range values {
			if strings.HasPrefix(value, query) {
				return true
			}
		}
	case "$":
		for _, value := range values {
			if strings.HasSuffix(value, query) {
				return true
			}
		}
	default:
		for _, value := range values {
			if value == query {
				return true
			}
		}
	}
	return false
}
