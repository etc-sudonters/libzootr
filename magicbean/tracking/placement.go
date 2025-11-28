package tracking

import (
	"sudonters/libzootr/components"
	"sudonters/libzootr/zecs"
)

type Placement struct {
	zecs.Proxy
	name components.Name
}

func (this Placement) DefaultToken(token Token) {
	this.Attach(components.DefaultPlacement(token.Entity()))
}

func (this Placement) Fixed(token Token) {
	this.Attach(components.HoldsToken(token.Entity()), components.Fixed{})
}

func (this Placement) Holding(token Token) {
	this.Attach(components.HoldsToken(token.Entity()))
}
