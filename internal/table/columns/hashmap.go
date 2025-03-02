package columns

import (
	"reflect"
	"sudonters/libzootr/internal/table"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

func NewMap() *Map {
	return NewMapWithCapacity(16)
}

func NewMapWithCapacity(capacity uint32) *Map {
	return &Map{
		entities: make(map[table.RowId]table.Value, capacity),
		members:  bitset32.WithBucketsFor(capacity),
	}
}

type Map struct {
	entities map[table.RowId]table.Value
	members  bitset32.Bitset
}

func (s *Map) Get(e table.RowId) table.Value {
	return s.entities[e]
}

func (s *Map) Set(e table.RowId, c table.Value) {
	s.entities[e] = c
	bitset32.Set(&s.members, e)
}

func (s *Map) Unset(e table.RowId) {
	delete(s.entities, e)
	bitset32.Unset(&s.members, e)
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

func (s *Map) Membership() bitset32.Bitset {
	return bitset32.Copy(s.members)
}

func HashMapColumn[T any](attr string) *table.ColumnBuilder {
	return table.BuildColumnOf[T](attr, NewMap())
}

func SizedHashMapColumn[T any](attr string, capacity uint32) *table.ColumnBuilder {
	return table.BuildColumnOf[T](attr, NewMapWithCapacity(capacity))
}
