package columns

import (
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"reflect"
	"sudonters/libzootr/table"
)

func NewMap() *Map {
	return NewMapWithCapacity(16)
}

func NewMapWithCapacity(capacity uint32) *Map {
	return &Map{make(map[table.RowId]table.Value, capacity)}
}

type Map struct {
	entities map[table.RowId]table.Value
}

func (s *Map) Get(e table.RowId) table.Value {
	return s.entities[e]
}

func (s *Map) Set(e table.RowId, c table.Value) {
	s.entities[e] = c
}

func (s *Map) Unset(e table.RowId) {
	delete(s.entities, e)
}

func (s *Map) ScanFor(v table.Value) (b bitset32.Bitset) {
	for id, value := range s.entities {
		if reflect.DeepEqual(v, value) {
			b.Set(uint32(id))
		}
	}

	return
}

func (s *Map) Len() int {
	return len(s.entities)
}

func HashMapColumn[T any]() *table.ColumnBuilder {
	return table.BuildColumnOf[T](NewMap())
}

func SizedHashMapColumn[T any](capacity uint32) *table.ColumnBuilder {
	return table.BuildColumnOf[T](NewMapWithCapacity(capacity))
}
