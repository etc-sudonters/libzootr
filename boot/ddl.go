package boot

import (
	"sudonters/libzootr/components"
	"sudonters/libzootr/internal/table"
	"sudonters/libzootr/internal/table/columns"
	"sudonters/libzootr/mido/symbols"
	"sudonters/libzootr/zecs"
)

func sizedslice[T zecs.Value](attr string, size uint32) zecs.DDL {
	return func() *table.ColumnBuilder {
		return columns.SizedSliceColumn[T](attr, size)
	}
}

func sizedbit[T zecs.Value](attr string, size uint32) zecs.DDL {
	return func() *table.ColumnBuilder {
		return columns.SizedBitColumnOf[T](attr, size)
	}
}

func sizedhash[T zecs.Value](attr string, capacity uint32) zecs.DDL {
	return func() *table.ColumnBuilder {
		return columns.SizedHashMapColumn[T](attr, capacity)
	}
}

func withattr(attr string, build func(string) *table.ColumnBuilder) zecs.DDL {
	return func() *table.ColumnBuilder {
		return build(attr)
	}
}

func staticddl() []zecs.DDL {
	return []zecs.DDL{
		sizedslice[components.Name]("/name", 9000),
		sizedbit[components.PlacementLocationMarker]("/world/placement", 5000),
		sizedhash[symbols.Kind]("/entity/kind", 5000),
		sizedhash[components.Ptr]("/entity/ptr", 5000),

		sizedhash[components.RuleParsed]("/logic/rule/parsed", 4000),
		sizedhash[components.RuleOptimized]("/logic/rule/optimized", 4000),
		sizedhash[components.RuleCompiled]("/logic/rule/compiled", 4000),
		sizedhash[components.EdgeKind]("/world/connection/kind", 4000),
		sizedhash[components.Connection]("/world/connection", 4000),
		sizedhash[components.RuleSource]("/logic/rule/source", 4000),
		sizedhash[components.DefaultPlacement]("/world/placement/default", 2200),

		withattr("/token/priority", columns.HashMapColumn[components.CollectablePriority]),
		withattr("/world/placement/holding", columns.HashMapColumn[components.HoldsToken]),
		withattr("/world/region/hint", columns.HashMapColumn[components.HintRegion]),
		withattr("/world/region/alt-hint", columns.HashMapColumn[components.AltHintRegion]),
		withattr("/world/dungeon/name", columns.HashMapColumn[components.DungeonName]),
		withattr("/world/location/savewarp", columns.HashMapColumn[components.Savewarp]),
		withattr("/world/scene", columns.HashMapColumn[components.Scene]),
		withattr("/logic/script/decl", columns.HashMapColumn[components.ScriptDecl]),
		withattr("/logic/script/source", columns.HashMapColumn[components.ScriptSource]),
		withattr("/logic/script/parsed", columns.HashMapColumn[components.ScriptParsed]),
		withattr("/name/alias", columns.HashMapColumn[components.AliasingName]),
		withattr("/token/kind/ocarina-note", columns.HashMapColumn[components.OcarinaNote]),
		withattr("/token/song/notes", columns.HashMapColumn[components.SongNotes]),
		withattr("/world/dungeon/group", columns.HashMapColumn[components.DungeonGroup]),
		withattr("/world/location/silver-rupee-puzzle", columns.HashMapColumn[components.SilverRupeePuzzle]),
		withattr("/token/song", columns.HashMapColumn[components.Song]),
		withattr("/token/price", columns.HashMapColumn[components.Price]),

		withattr("/world/location/skipped", columns.BitColumnOf[components.Skipped]),
		withattr("/world/location/collected", columns.BitColumnOf[components.Collected]),
		withattr("/token", columns.BitColumnOf[components.TokenMarker]),
		withattr("/world/region", columns.BitColumnOf[components.RegionMarker]),
		withattr("/world/is-boss-room", columns.BitColumnOf[components.IsBossRoom]),
		withattr("/world/time-passes", columns.BitColumnOf[components.TimePassess]),
		withattr("/token/kind/compass", columns.BitColumnOf[components.Compass]),
		withattr("/token/kind/drop", columns.BitColumnOf[components.Drop]),
		withattr("/token/kind/reward", columns.BitColumnOf[components.DungeonReward]),
		withattr("/token/kind/event", columns.BitColumnOf[components.Event]),
		withattr("/token/kind/item", columns.BitColumnOf[components.Item]),
		withattr("/token/kind/map", columns.BitColumnOf[components.Map]),
		withattr("/token/kind/refill", columns.BitColumnOf[components.Refill]),
		withattr("/world/location/shop", columns.BitColumnOf[components.Shop]),
		withattr("/token/kind/silver-rupee", columns.BitColumnOf[components.SilverRupee]),
		withattr("/token/kind/silver-rupee-pouch", columns.BitColumnOf[components.SilverRupeePouch]),
		withattr("/token/kind/small-key", columns.BitColumnOf[components.SmallKey]),
		withattr("/token/kind/boss-key", columns.BitColumnOf[components.BossKey]),
		withattr("/token/kind/key-ring", columns.BitColumnOf[components.DungeonKeyRing]),
		withattr("/token/kind/gold-skull-token", columns.BitColumnOf[components.GoldSkulltulaToken]),
		withattr("/token/kind/medallion", columns.BitColumnOf[components.Medallion]),
		withattr("/token/kind/stone", columns.BitColumnOf[components.Stone]),
		withattr("/token/kind/bottle", columns.BitColumnOf[components.Bottle]),
		withattr("/world/location/root", columns.BitColumnOf[components.WorldGraphRoot]),
	}
}
