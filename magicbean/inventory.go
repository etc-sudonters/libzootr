package magicbean

import (
	"sudonters/libzootr/zecs"
)

func EmptyInventory() Inventory {
	return Inventory(make(map[zecs.Entity]int))
}

func CopyInventory(i Inventory) Inventory {
	copy := make(Inventory, len(i))

	for k, v := range i {
		copy[k] = v
	}
	return copy
}

func DiffInventories(old, new Inventory) Inventory {
	diff := make(Inventory, len(new))

	for k := range new {
		had := old[k]
		has := new[k]
		if has-had == 0 {
			continue
		}
		diff[k] = has - had
	}

	return diff
}

type Inventory map[zecs.Entity]int

func (this Inventory) CollectOne(entity zecs.Entity) {
	this.Collect(entity, 1)
}

func (this Inventory) Collect(entity zecs.Entity, n int) {
	this[entity] += n
}

func (this Inventory) CollectOneEach(entities []zecs.Entity) {
	for _, entity := range entities {
		this.Collect(entity, 1)
	}
}

func (this Inventory) Remove(entity zecs.Entity, n int) int {
	has := this[entity]

	switch {
	case has == 0:
		return 0
	case n == has:
		delete(this, entity)
		return n
	case n < has:
		this[entity] = has - n
		return n
	case n > has:
		delete(this, entity)
		return has
	default:
		panic("unreachable")
	}
}

func (this Inventory) Count(entity zecs.Entity) int {
	return this[entity]
}

func (this Inventory) Sum(entities []zecs.Entity) int {
	var total int

	for _, entity := range entities {
		total += this[entity]
	}

	return total
}

func (this Inventory) AddFrom(new Inventory) {
	for item, qty := range new {
		has := this[item]
		this[item] = has + qty
	}
}
