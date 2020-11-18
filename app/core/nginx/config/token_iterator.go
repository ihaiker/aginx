package config

import "fmt"

type tokenIterator struct {
	it  *charIterator
	opt *Options
}

func newTokenIterator(filename string, opt *Options) *tokenIterator {
	return &tokenIterator{it: newCharIterator(filename), opt: opt}
}

func newTokenIteratorWithBytes(bs []byte, opt *Options) *tokenIterator {
	return &tokenIterator{it: newCharIteratorWithBytes(bs), opt: opt}
}

func (self *tokenIterator) next() (token string, tokenLine int, tokenHas bool) {
	for {
		char, line, has := self.it.nextFilter(ValidChars)
		if !has {
			return
		}
		switch char {
		case ";", "{", "}":
			{
				token = char
				tokenLine = line
				tokenHas = true
				return
			}
		case "#":
			{
				word, _, _ := self.it.nextTo(In("\n"), false)
				token = char + word
				tokenLine = line
				tokenHas = true
				return
			}
		case "'", `"`, "`":
			{
				word, _, wordHas := self.it.nextTo(In(char), true)
				if !wordHas {
					panic(fmt.Errorf("error at line : %d", line))
				}
				if self.opt.RemoveBrackets {
					token = word[0 : len(word)-1] //去除文本括号
				} else {
					token = char + word
				}
				tokenLine = line
				tokenHas = true
				return
			}
		default:
			word, _, wordHas := self.it.nextTo(Not(ValidChars).Or(In(";", "{")), false)
			if !wordHas {
				panic(fmt.Errorf("error at line : %d", line))
			}
			token = char + word
			tokenLine = line
			tokenHas = true
			return
		}
	}
}
