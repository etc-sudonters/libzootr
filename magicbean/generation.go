package magicbean

import (
	"math/rand/v2"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/settings"

	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/zecs"
)

type Generation struct {
	Ocm       zecs.Ocm
	World     ExplorableWorld
	Objects   objects.Table
	Inventory Inventory
	Rng       rand.Rand
	Settings  *settings.Model
	Tokens    tracking.Tokens
	Nodes     tracking.Nodes
}
