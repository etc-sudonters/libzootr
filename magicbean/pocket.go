package magicbean

import (
	"maps"
	"slices"
	"sudonters/libzootr/components"
	"sudonters/libzootr/zecs"
)

func NewPockets(inventory Inventory, ocm *zecs.Ocm) Pocket {
	var pocket Pocket
	pocket.inventory = inventory
	pocket.heartPiece = zecs.FindOne(ocm, components.Name("Piece of Heart"), zecs.With[components.TokenMarker])
	pocket.scarecrowSong = zecs.FindOne(ocm, components.Name("Scarecrow Song"), zecs.With[components.TokenMarker])
	pocket.transcribe = zecs.IndexEntities[components.OcarinaNote](ocm)
	pocket.songs = zecs.IndexValue[components.SongNotes](ocm)
	pocket.bottles = zecs.SliceMatching(ocm, zecs.With[components.Bottle])
	pocket.stones = zecs.SliceMatching(ocm, zecs.With[components.Stone])
	pocket.meds = zecs.SliceMatching(ocm, zecs.With[components.Medallion])
	pocket.rewards = zecs.SliceMatching(ocm, zecs.With[components.DungeonReward])
	pocket.notes = slices.Collect(maps.Values(pocket.transcribe))
	return pocket
}

type Pocket struct {
	inventory  Inventory
	transcribe map[components.OcarinaNote]zecs.Entity
	songs      map[zecs.Entity]components.SongNotes

	heartPiece, scarecrowSong             zecs.Entity
	bottles, stones, meds, rewards, notes []zecs.Entity
}

func (this Pocket) Has(entity zecs.Entity, n int) bool {
	return this.inventory.Count(entity) >= n
}

func (this Pocket) HasEvery(entities []zecs.Entity) bool {
	for _, entity := range entities {
		if !this.Has(entity, 1) {
			return false
		}
	}
	return true
}

func (this Pocket) HasAny(entities []zecs.Entity) bool {
	for _, entity := range entities {
		if this.Has(entity, 1) {
			return true
		}
	}
	return false
}

func (this Pocket) HasBottle() bool {
	return this.HasAny(this.bottles)
}

func (this Pocket) HasStones(n int) bool {
	return this.inventory.Sum(this.stones) >= n
}

func (this Pocket) HasMedallions(n int) bool {
	return this.inventory.Sum(this.meds) >= n
}

func (this Pocket) HasDungeonRewards(n int) bool {
	return this.inventory.Sum(this.rewards) >= n
}

func (this Pocket) HasHearts(n float64) bool {
	pieces := float64(this.inventory.Count(this.heartPiece))
	hearts := pieces / 4
	return hearts >= n
}

func (this Pocket) HasAllNotes(entity zecs.Entity) bool {
	if entity == this.scarecrowSong {
		return this.inventory.Sum(this.notes) >= 2
	}

	song, exists := this.songs[entity]
	if !exists {
		panic("not a song!")
	}
	notes := []components.OcarinaNote(song)
	transcript := make([]zecs.Entity, len(notes))
	for i, note := range notes {
		entity, exists := this.transcribe[note]
		if !exists {
			panic("unknown note")
		}
		transcript[i] = entity
	}

	return this.HasEvery(transcript)
}
