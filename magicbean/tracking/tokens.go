package tracking

import (
	"sudonters/libzootr/components"
	"sudonters/libzootr/zecs"
)

func NewTokens(ocm *zecs.Ocm) Tokens {
	return Tokens{named[components.TokenMarker](ocm), ocm}
}

type Tokens struct {
	tokens namedents
	parent *zecs.Ocm
}

type Token struct {
	zecs.Proxy
	name name
}

func (this Tokens) Named(name name) Token {
	token := this.tokens.For(name)
	token.Attach(components.TokenMarker{})
	return Token{token, name}
}

func (this Tokens) MustGet(name name) Token {
	token := this.tokens.MustGet(name)
	return Token{token, name}
}
