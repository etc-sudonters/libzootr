package tracking

import (
	"sudonters/libzootr/internal"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/ocm"

	"github.com/etc-sudonters/substrate/slipup"
)

type namedents = Tracked[magicbean.Name]
type directed = magicbean.Connection
type name = magicbean.Name

var namef = magicbean.NameF

func named[T ocm.Component](entities *ocm.Entities) (namedents, error) {
	return Track[name](entities, table.Exists[T])
}

func NewNodes(entities *ocm.Entities) (Nodes, error) {
	var this Nodes
	var err error
	if this.regions, err = named[magicbean.Region](entities); err != nil {
		return this, slipup.Describe(err, "failed to retrieve regions")
	}
	if this.placements, err = named[magicbean.Placement](entities); err != nil {
		return this, slipup.Describe(err, "failed to retrieve placements")
	}
	if this.transit, err = Track[directed](entities); err != nil {
		return this, slipup.Describe(err, "failed to retrieve connections")
	}
	this.parent = entities
	return this, nil
}

type Nodes struct {
	regions, placements namedents
	transit             Tracked[directed]
	parent              *ocm.Entities
}

type Region struct {
	ocm.Proxy
	name   name
	parent Nodes
}

type Transit struct {
	ocm.Proxy
	name   name
	t      directed
	parent Nodes
}

func (this Nodes) Region(name name) Region {
	region, err := this.regions.For(name)
	internal.PanicOnError(err)
	internal.PanicOnError(region.Attach(magicbean.Region{}))
	return Region{region, name, this}
}

func (this Nodes) Placement(name name) Placement {
	place, err := this.placements.For(name)
	internal.PanicOnError(err)
	internal.PanicOnError(place.Attach(magicbean.Placement{}))
	return Placement{place, name}
}

func (this Region) ConnectsTo(other Region) Transit {
	return this.connect(other.Entity(), other.name, magicbean.EdgeTransit)
}

func (this Region) Has(node Placement) Transit {
	return this.connect(node.Entity(), node.name, magicbean.EdgePlacement)
}

func (this Region) connect(to ocm.Entity, toName name, kind magicbean.EdgeKind) Transit {
	name := namef("%s -> %s", this.name, toName)
	directed := directed{From: this.Entity(), To: to}
	proxy, err := this.parent.transit.For(directed)
	internal.PanicOnError(err)
	transit := Transit{
		Proxy:  proxy,
		name:   name,
		t:      directed,
		parent: this.parent,
	}
	transit.Proxy.Attach(name, kind)
	return transit
}
