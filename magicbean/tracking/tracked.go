package tracking

import (
	"fmt"
	"slices"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/ocm"
)

func Track[K ocm.ComparableComponent](entities *ocm.Entities, qs ...table.Q) (Tracked[K], error) {
	q := []table.Q{table.Load[K]}
	q = slices.Concat(q, qs)
	rows, err := entities.Query(q...)
	if err != nil {
		return Tracked[K]{}, err
	}

	track := Tracked[K]{
		parent: entities,
		cache:  make(map[K]ocm.Entity, rows.Len()),
	}

	for row, tup := range rows.All {
		k := tup.Values[0].(K)
		track.cache[k] = row
	}

	return track, nil
}

type Tracked[K ocm.ComparableComponent] struct {
	parent *ocm.Entities
	cache  map[K]ocm.Entity
}

func (this *Tracked[K]) For(key K) (ocm.Proxy, error) {
	if entity, exists := this.cache[key]; exists {
		proxy, _ := this.parent.Proxy(entity)
		return proxy, nil
	}

	entity, err := this.parent.CreateEntity()
	if err != nil {
		return entity, err
	}
	this.cache[key] = entity.Entity()
	if err := entity.Attach(key); err != nil {
		return entity, err
	}

	return entity, nil
}

func (this *Tracked[K]) MustGet(key K) ocm.Proxy {
	entity, exists := this.cache[key]
	if !exists {
		panic(fmt.Errorf("no entity registered for key %#v", key))
	}

	proxy, err := this.parent.Proxy(entity)
	internal.PanicOnError(err)
	return proxy
}
