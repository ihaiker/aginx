package config

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"os"
	"strings"
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

func subDirectives(it *tokenIterator, opt *Options) ([]*Directive, error) {
	directives := make([]*Directive, 0)
	for {
		token, line, has := it.next()
		if !has {
			break
		}
		if token == ";" || token == "}" {
			break
		} else if token[0] == '#' { //注释
			if !opt.RemoveAnnotation {
				directives = append(directives, &Directive{
					Line: line, Name: "#", Args: []string{strings.Trim(token[1:], " ")},
				})
			}
		} else {
			if args, lastToken, err := expectNextToken(it, In(";", "{")); err != nil {
				return nil, err
			} else if lastToken == ";" {
				if opt.Delimiter {
					if strings.HasSuffix(token, ":") {
						token = token[0 : len(token)-1]
					}
				}
				directives = append(directives, &Directive{
					Line: line, Name: token, Args: args,
				})
			} else {
				if opt.Delimiter {
					if strings.HasSuffix(token, ":") {
						token = token[0 : len(token)-1]
					}
				}
				directive := &Directive{
					Line: line, Name: token, Args: args,
				}
				if subDirs, err := subDirectives(it, opt); err != nil {
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

func MustParse(filename string, opt *Options) *Configuration {
	cfg, err := Parse(filename, opt)
	errors.PanicMessage(err, "parse "+filename)
	return cfg
}

func Parse(filename string, opt *Options) (cfg *Configuration, err error) {
	if opt == nil {
		opt = &Options{
			Delimiter:        false,
			RemoveBrackets:   false,
			RemoveAnnotation: false,
		}
	}
	defer errors.Catch(func(re error) {
		err = re
	})
	cfg = &Configuration{Name: filename}
	it := newTokenIterator(filename, opt)
	cfg.Body, err = subDirectives(it, opt)
	return
}

func ParseWith(filename string, bs []byte, opt *Options) (cfg *Configuration, err error) {
	defer errors.Catch(func(re error) {
		err = re
	})
	if opt == nil {
		opt = &Options{
			Delimiter:        false,
			RemoveBrackets:   false,
			RemoveAnnotation: false,
		}
	}
	cfg = &Configuration{Name: filename}
	it := newTokenIteratorWithBytes(bs, opt)
	cfg.Body, err = subDirectives(it, opt)
	return
}
