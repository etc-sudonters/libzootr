package magicbean

import (
	"math/rand/v2"
	"sudonters/libzootr/internal/settings"
	"sudonters/libzootr/table/ocm"

	"sudonters/libzootr/mido/objects"
)

type Generation struct {
	Entities  *ocm.Entities
	World     ExplorableWorld
	Objects   objects.Table
	Inventory Inventory
	Rng       rand.Rand
	Settings  settings.Zootr
}
