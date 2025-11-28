package tracking

import (
	"sudonters/libzootr/components"
	"sudonters/libzootr/zecs"
)

type namedents = zecs.Tracked[components.Name]
type directed = components.Connection
type name = components.Name

var namef = components.NameF

func named[T zecs.Value](ocm *zecs.Ocm) namedents {
	return zecs.Tracking[name](ocm, zecs.With[T])
}

func NewNodes(ocm *zecs.Ocm) Nodes {
	var this Nodes
	this.regions = named[components.RegionMarker](ocm)
	this.placements = named[components.PlacementLocationMarker](ocm)
	this.transit = zecs.Tracking[directed](ocm)
	this.parent = ocm
	return this
}

type Nodes struct {
	regions, placements namedents
	transit             zecs.Tracked[directed]
	parent              *zecs.Ocm
}

type Region struct {
	zecs.Proxy
	name   name
	parent Nodes
}

type Transit struct {
	zecs.Proxy
	name   name
	t      directed
	parent Nodes
}

func (this Nodes) Region(name name) Region {
	region := this.regions.For(name)
	region.Attach(components.RegionMarker{})
	return Region{region, name, this}
}

func (this Nodes) Placement(name name) Placement {
	place := this.placements.For(name)
	place.Attach(components.PlacementLocationMarker{})
	return Placement{place, name}
}

func (this Region) ConnectsTo(other Region) Transit {
	return this.connect(other.Entity(), other.name, components.EdgeTransit)
}

func (this Region) Has(node Placement) Transit {
	return this.connect(node.Entity(), node.name, components.EdgePlacement)
}

func (this Region) connect(to zecs.Entity, toName name, kind components.EdgeKind) Transit {
	name := namef("%s -> %s", this.name, toName)
	directed := directed{From: this.Entity(), To: to}
	transit := Transit{
		Proxy:  this.parent.transit.For(directed),
		name:   name,
		t:      directed,
		parent: this.parent,
	}
	transit.Proxy.Attach(name, kind)
	return transit
}
