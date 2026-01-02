package tracking

import (
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/table/ocm"
)

func NewTokens(ocm *ocm.Entities) (Tokens, error) {
	tokens, err := named[magicbean.Token](ocm)
	return Tokens{tokens, ocm}, err
}

type Tokens struct {
	tokens namedents
	parent *ocm.Entities
}

type Token struct {
	ocm.Proxy
	name name
}

func (this Tokens) Named(name name) (Token, error) {
	token, err := this.tokens.For(name)
	if err != nil {
		return Token{}, err
	}
	err = token.Attach(magicbean.Token{})
	return Token{token, name}, err
}

func (this Tokens) MustGet(name name) Token {
	token := this.tokens.MustGet(name)
	return Token{token, name}
}
