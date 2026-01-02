package magicbean

import (
	"maps"
	"slices"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/ocm"
)

func NewPockets(inventory *Inventory, entities *ocm.Entities) Pocket {
	var pocket Pocket
	pocket.inventory = inventory
	{
		heartPiece, heartErr := ocm.FindOne(entities, Name("Piece of Heart"), table.Exists[Token])
		internal.PanicOnError(heartErr)
		pocket.heartPiece = heartPiece
	}

	{
		scarecrowSong, scarecrowErr := ocm.FindOne(entities, Name("Scarecrow Song"), table.Exists[Token])
		internal.PanicOnError(scarecrowErr)
		pocket.scarecrowSong = scarecrowSong
	}

	{
		transcribe, transcribeErr := ocm.KeyedEntities[OcarinaNote](entities)
		internal.PanicOnError(transcribeErr)
		pocket.transcribe = maps.Collect(transcribe)
	}

	{
		songs, songErr := ocm.IndexedComponent[SongNotes](entities)
		internal.PanicOnError(songErr)
		pocket.songs = maps.Collect(songs)
	}

	{
		bottles, bottleErr := entities.Matching(table.Exists[Bottle])
		internal.PanicOnError(bottleErr)
		pocket.bottles = slices.Collect(bottles)
	}

	{
		stones, stoneErr := entities.Matching(table.Exists[Stone])
		internal.PanicOnError(stoneErr)
		pocket.stones = slices.Collect(stones)
	}

	{
		meds, medErr := entities.Matching(table.Exists[Medallion])
		internal.PanicOnError(medErr)
		pocket.meds = slices.Collect(meds)
	}

	{
		rewards, rewardErr := entities.Matching(table.Exists[DungeonReward])
		internal.PanicOnError(rewardErr)
		pocket.rewards = slices.Collect(rewards)
	}

	pocket.notes = slices.Collect(maps.Values(pocket.transcribe))
	return pocket
}

type Pocket struct {
	inventory  *Inventory
	transcribe map[OcarinaNote]ocm.Entity
	songs      map[ocm.Entity]SongNotes

	heartPiece, scarecrowSong             ocm.Entity
	bottles, stones, meds, rewards, notes []ocm.Entity
}

func (this Pocket) Has(entity ocm.Entity, n float64) bool {
	return this.inventory.Count(entity) >= n
}

func (this Pocket) HasEvery(entities []ocm.Entity) bool {
	for _, entity := range entities {
		if !this.Has(entity, 1) {
			return false
		}
	}
	return true
}

func (this Pocket) HasAny(entities []ocm.Entity) bool {
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

func (this Pocket) HasStones(n float64) bool {
	return this.inventory.Sum(this.stones) >= n
}

func (this Pocket) HasMedallions(n float64) bool {
	return this.inventory.Sum(this.meds) >= n
}

func (this Pocket) HasDungeonRewards(n float64) bool {
	return this.inventory.Sum(this.rewards) >= n
}

func (this Pocket) HasHearts(n float64) bool {
	pieces := this.inventory.Count(this.heartPiece)
	hearts := pieces / 4
	return hearts >= n
}

func (this Pocket) HasAllNotes(entity ocm.Entity) bool {
	if entity == this.scarecrowSong {
		return this.inventory.Sum(this.notes) >= 2
	}

	song, exists := this.songs[entity]
	if !exists {
		panic("not a song!")
	}
	notes := []OcarinaNote(song)
	transcript := make([]ocm.Entity, len(notes))
	for i, note := range notes {
		entity, exists := this.transcribe[note]
		if !exists {
			panic("unknown note")
		}
		transcript[i] = entity
	}

	return this.HasEvery(transcript)
}
