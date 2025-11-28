package magicbean

import (
	"fmt"
	"sudonters/libzootr/components"
	"sudonters/libzootr/settings"
	"sudonters/libzootr/zecs"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

func CollectTokensFrom(
	ocm *zecs.Ocm,
	from bitset32.Bitset,
	inventory Inventory,
) error {
	q := ocm.Query()
	q.Build(
		zecs.Load[components.HoldsToken],
		zecs.With[components.PlacementLocationMarker],
		zecs.WithOut[components.Collected],
		zecs.FromSubset(&from),
	)

	rows, err := q.Execute()
	if err != nil {
		return err
	}

	for entity, tup := range rows.All {
		token := tup.Values[0].(components.HoldsToken)
		inventory.CollectOne(zecs.Entity(token))

		proxy := ocm.Proxy(entity)
		err := proxy.Attach(components.Collected{})
		if err != nil {
			return fmt.Errorf("while marking %v as collected: %w", entity, err)
		}
	}

	return nil
}

func CollectStartingItems(generation *Generation, additional ...components.Name) error {
	tokens := &generation.Tokens
	these := generation.Settings
	inventory := generation.Inventory

	for name, qty := range these.Logic.Spawns.Items {
		token := tokens.MustGet(components.Name(name))
		inventory.Collect(token.Entity(), qty)
	}

	if these.Logic.Shuffling.Flags&settings.ShuffleOcarinaNotes != settings.ShuffleOcarinaNotes {
		buttons := []zecs.Entity{
			tokens.MustGet("Ocarina A Button").Entity(),
			tokens.MustGet("Ocarina C left Button").Entity(),
			tokens.MustGet("Ocarina C right Button").Entity(),
			tokens.MustGet("Ocarina C up Button").Entity(),
			tokens.MustGet("Ocarina C down Button").Entity(),
		}
		inventory.CollectOneEach(buttons)
	}

	for _, name := range additional {
		inventory.CollectOne(tokens.MustGet(name).Entity())
	}

	ocm := &generation.Ocm
	skipped := zecs.Bitset32Matching(ocm, zecs.With[components.Skipped])
	return CollectTokensFrom(ocm, skipped, inventory)
}
