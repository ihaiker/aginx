package config

import "fmt"

type tokenIterator struct {
	it    charIterator
	token string
	line  int
}

func newTokenIterator(filename string) tokenIterator {
	return tokenIterator{it: newCharIterator(filename)}
}

func newTokenIteratorWithBytes(bs []byte) tokenIterator {
	return tokenIterator{it: newCharIteratorWithBytes(bs)}
}

func (self *tokenIterator) next() (token string, tokenLine int, tokenHas bool) {
	if self.token != "" {
		token = self.token
		tokenLine = self.line
		tokenHas = true

		self.token = ""
		return
	}
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
				word, _, _ := self.it.nextTo(In("\n"), UnIncludeLastChar)
				token = char + word
				tokenLine = line
				tokenHas = true
				return
			}
		case "'", `"`:
			{
				word, _, wordHas := self.it.nextTo(In(char), IncludeLastChar)
				if !wordHas {
					panic(fmt.Errorf("error at line : %d", line))
				}
				token = char + word
				tokenLine = line
				tokenHas = true
				return
			}
		default:
			word, _, wordHas := self.it.nextTo(Not(ValidChars).Or(In(";", "{")), UnIncludeLastChar)
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
