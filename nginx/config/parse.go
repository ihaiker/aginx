package config

import (
	"fmt"
	"github.com/ihaiker/aginx/util"
	"os"
)

func expectNextToken(it *tokenIterator, filter CharFilter) ([]string, string, error) {
	tokens := make([]string, 0)
	for {
		if token, _, has := it.next(); has {
			if filter(token, "") {
				return tokens, token, nil
			}
			tokens = append(tokens, token)
		} else {
			return nil, "", os.ErrNotExist
		}
	}
}

func subDirectives(it *tokenIterator) ([]*Directive, error) {
	directives := make([]*Directive, 0)
	for {
		token, line, has := it.next()
		if !has {
			break
		}
		if token == ";" || token == "}" {
			break
		} else if token[0] == '#' { //注释
			directives = append(directives, &Directive{
				Line: line, Name: "#", Args: []string{token},
			})
		} else {
			if args, lastToken, err := expectNextToken(it, In(";", "{")); err != nil {
				return nil, err
			} else if lastToken == ";" {
				directives = append(directives, &Directive{
					Line: line, Name: token, Args: args,
				})
			} else {
				directive := &Directive{
					Line: line, Name: token, Args: args,
				}
				if subDirs, err := subDirectives(it); err != nil {
					return nil, fmt.Errorf("line %d, %s ", line, token)
				} else {
					directive.Body = subDirs
				}
				directives = append(directives, directive)
			}
		}
	}
	return directives, nil
}

func MustParse(filename string) *Configuration {
	cfg, err := Parse(filename)
	util.PanicMessage(err, "parse "+filename)
	return cfg
}

func Parse(filename string) (cfg *Configuration, err error) {
	defer util.Catch(func(re error) {
		err = re
	})
	cfg = &Configuration{Name: filename}
	it := newTokenIterator(filename)
	cfg.Body, err = subDirectives(it)
	return
}

func ParseWith(filename string, bs []byte) (cfg *Configuration, err error) {
	defer util.Catch(func(re error) {
		err = re
	})
	cfg = &Configuration{Name: filename}
	it := newTokenIteratorWithBytes(bs)
	cfg.Body, err = subDirectives(it)
	return
}
